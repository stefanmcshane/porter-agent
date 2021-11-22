package processor

import (
	"context"

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
		redisClient:  redis.NewClient(redisHost, redisPort, "", "", redis.NODESTORE, maxTailLines),
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

// to trigger actual request for porter server in case of
// a Delete or Failed/Unknown Phase
func (n *NodeEventProcessor) AddToWorkQueue(context context.Context, object types.NamespacedName, details *models.EventDetails) {
	panic("not implemented") // TODO: Implement
}
