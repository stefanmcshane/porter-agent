package processor

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/porter-dev/porter-agent/pkg/models"
	"github.com/porter-dev/porter-agent/pkg/redis"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type NodeEventProcessor struct {
	redisClient  *redis.Client
	resourceType models.EventResourceType

	logger logr.Logger
}

func NewNodeEventProcessor() Interface {
	return &NodeEventProcessor{
		resourceType: models.NodeResource,
		redisClient:  redis.NewClient(redisHost, redisPort, "", "", redis.PODSTORE, maxTailLines),
		logger:       log.Log.WithName("node-event-processor"),
	}
}

// EnqueueDetails Used in case of normal events to store and update logs
func (n *NodeEventProcessor) EnqueueDetails(context context.Context, object types.NamespacedName, options *EnqueueDetailOptions) {
	instance := options.NodeInstance

	details := models.MarshallableNodeConditions(instance.Status.Conditions)
	n.logger.Info("appending node details")
	err := n.redisClient.AppendAndTrimDetails(context, models.NodeResource, object.Namespace, object.Name, details)
	if err != nil {
		n.logger.Error(err, "error encountered while appending details to redis")
	}
}

// AddToWorkQueue to trigger actual request for porter server in case of
// a Delete or Failed/Unknown Phase
func (n *NodeEventProcessor) AddToWorkQueue(context context.Context, object types.NamespacedName, details *models.EventDetails) {
	logger := log.Log.WithName("node-event-processor")

	exists, err := n.redisClient.ErroredItemExists(context, models.NodeResource, "", object.Name)
	if err != nil {
		logger.Error(err, "unable to check item's existence in error register")
		return
	}

	if exists {
		logger.Info("pushing to work queue")
		err := n.pushToWorkQueue(context, details)
		if err != nil {
			logger.Error(err, "unable to push item to work queue")
		}

		if !details.Critical {
			logger.Info("deleting from error register")
			err := n.redisClient.DeleteErroredItem(context, n.resourceType, object.Namespace, object.Name)
			if err != nil {
				logger.Error(err, "unable to register errored item")
			}
		}
	} else if details.Critical {
		logger.Info("adding to error register")
		err := n.redisClient.RegisterErroredItem(context, n.resourceType, object.Namespace, object.Name)
		if err != nil {
			logger.Error(err, "unable to register errored item")
		}

		logger.Info("pushing to work queue")
		err = n.pushToWorkQueue(context, details)
		if err != nil {
			logger.Error(err, "unable to push item to work queue")
		}
	}
}

func (n *NodeEventProcessor) pushToWorkQueue(ctx context.Context, details *models.EventDetails) error {
	packed, err := json.Marshal(details)
	if err != nil {
		return err
	}

	err = n.redisClient.AppendToNotifyWorkQueue(ctx, packed)
	if err != nil {
		return err
	}

	return nil
}
