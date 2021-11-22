package processor

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/porter-dev/porter-agent/pkg/models"
	"github.com/porter-dev/porter-agent/pkg/redis"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PodEventProcessor is the pod processor that holds
// a kube clientset, a redis client and the resource type
// it is associated with
type PodEventProcessor struct {
	kubeClient   *kubernetes.Clientset
	redisClient  *redis.Client
	resourceType models.EventResourceType
}

// NewPodEventProcessor returns a pod processor
// with an instance of a kubernetes clientset
// created with a given config
func NewPodEventProcessor(config *rest.Config) Interface {
	return &PodEventProcessor{
		kubeClient:   kubernetes.NewForConfigOrDie(config),
		resourceType: models.PodResource,
		redisClient:  redis.NewClient(redisHost, redisPort, "", "", redis.PODSTORE, maxTailLines),
	}
}

// EnqueueDetails is used in case of normal events to store and update logs
func (p *PodEventProcessor) EnqueueDetails(ctx context.Context, object types.NamespacedName, options *EnqueueDetailOptions) {
	logger := log.Log.WithName("event-processor")

	logOptions := &corev1.PodLogOptions{
		TailLines: &maxTailLines,
	}

	options.SetContainerName(logOptions)

	req := p.kubeClient.
		CoreV1().
		Pods(object.Namespace).
		GetLogs(object.Name, logOptions)

	podLogs, err := req.Stream(ctx)
	if err != nil {
		logger.Error(err, "error streaming logs")
		return
	}
	defer podLogs.Close()

	logs := new(bytes.Buffer)
	_, err = io.Copy(logs, podLogs)
	if err != nil {
		logger.Error(err, "unable to read logs")
		return
	}

	strLogs := logs.String()
	logger.Info("Successfully fetched logs", "object", object)

	// update logs in the redis store
	err = p.redisClient.AppendAndTrimDetails(ctx, p.resourceType, object.Namespace, object.Name, strings.Split(strLogs, "\n"))
	if err != nil {
		logger.Error(err, "unable to append logs to the store")
		return
	}
}

// AddToWorkQueue is supposed to check if the event's object is
// present in error register, here we might have the following cases:
//	1. object in the error register
//		- current event is non critical
//			- push to work queue as this means transition from error to healthy state
//			- remove from error register as this should be marked as healthy now
//		- current event is critical
//			- push to work queue as its the repeat occurance of the same error
//			  however this can be turned off in future to significantly reduce the
//			  the repeated events of long errored pods
//	2. object not in error register
//		- current event is not critical
//			- don't push the event to work queue, just let it be used for log refresh
//		- current event is critical
//			- add to register
//			- push to work queue
// the relevant event in a work queue
func (p *PodEventProcessor) AddToWorkQueue(ctx context.Context, object types.NamespacedName, details *models.EventDetails) {
	logger := log.Log.WithName("event-processor")

	logger.Info("current pod condition", "details", details)

	exists, err := p.redisClient.ErroredItemExists(ctx, p.resourceType, object.Namespace, object.Name)
	if err != nil {
		logger.Error(err, "unable to check items existence in error register")
	}

	if exists {
		logger.Info("pushing to work queue")
		err := p.pushToWorkQueue(ctx, details)
		if err != nil {
			logger.Error(err, "unable to push items to work queue")
		}

		if !details.Critical {
			// delete from error register
			logger.Info("deleting from error register")
			err := p.redisClient.DeleteErroredItem(ctx, p.resourceType, object.Namespace, object.Name)
			if err != nil {
				logger.Error(err, "unable to delete item from error register")
			}
		}
	} else if details.Critical {
		logger.Info("adding to error register")
		err := p.redisClient.RegisterErroredItem(ctx, p.resourceType, object.Namespace, object.Name)
		if err != nil {
			logger.Error(err, "unable to register errored item")
		}

		logger.Info("pushing to work queue")
		err = p.pushToWorkQueue(ctx, details)
		if err != nil {
			logger.Error(err, "unable to push item to work queue")
		}
	}
}

func (p *PodEventProcessor) pushToWorkQueue(ctx context.Context, details *models.EventDetails) error {
	packed, err := json.Marshal(details)
	if err != nil {
		return err
	}

	err = p.redisClient.AppendToNotifyWorkQueue(ctx, packed)
	if err != nil {
		return err
	}

	return nil
}
