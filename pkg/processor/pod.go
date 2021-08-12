package processor

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/porter-dev/porter-agent/pkg/models"
	"github.com/porter-dev/porter-agent/pkg/redis"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	redisHost    string
	redisPort    string
	maxTailLines int64
)

func init() {
	viper.SetDefault("REDIS_HOST", "porter-redis")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("MAX_TAIL_LINES", int64(100))
	viper.AutomaticEnv()

	redisHost = viper.GetString("REDIS_HOST")
	redisPort = viper.GetString("REDIS_PORT")
	maxTailLines = viper.GetInt64("MAX_TAIL_LINES")
}

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

// EnqueueWithLogLines is used in case of normal events to store and update logs
func (p *PodEventProcessor) EnqueueWithLogLines(ctx context.Context, object types.NamespacedName) {
	logger := log.FromContext(ctx)

	req := p.kubeClient.
		CoreV1().
		Pods(object.Namespace).
		GetLogs(object.Name, &corev1.PodLogOptions{
			TailLines: &maxTailLines,
		})

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
	logger.Info("Successfully fetched logs")
	// update logs in the redis store
	err = p.redisClient.AppendAndTrimDetails(ctx, p.resourceType.String(), object.Namespace, object.Name, strings.Split(strLogs, "\n"))
	if err != nil {
		logger.Error(err, "unable to append logs to the store")
		return
	}
}

// TriggerNotifyForEvent is supposed to trigger actual
// request for porter server in case of a Delete or
// Failed/Unknown Phase over HTTP. If that fails, it stores
// the relevant event in a work queue
func (p *PodEventProcessor) TriggerNotifyForEvent(ctx context.Context, object types.NamespacedName, details models.EventDetails) {
	logger := log.FromContext(ctx)
	logger.Info("notification triggered")

	logger.Info("current pod condition", "details", details)

	// call HTTP client and try posting on the porter server
	// in case of failure, append to the NotifyWorkQueue in redis

	// TODO: implement/call HTTP layer

	// assume HTTP failed, push to redis work queue
	packed, err := json.Marshal(details)
	if err != nil {
		logger.Error(err, "unable to marshal details to a json object")
		return
	}

	err = p.redisClient.AppendToNotifyWorkQueue(ctx, packed)
	if err != nil {
		logger.Error(err, "unable to push notify to work queue")
		return
	}
}
