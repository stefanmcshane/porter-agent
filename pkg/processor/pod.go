package processor

import (
	"context"

	"github.com/porter-dev/porter-agent/pkg/models"
	"github.com/porter-dev/porter-agent/pkg/redis"
	"github.com/spf13/viper"
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
	viper.SetDefault("REDIS_HOST", "porter-redis-master")
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

func (p *PodEventProcessor) NotifyNewIncident(ctx context.Context, incidentID string) {
	logger := log.Log.WithName("event-processor")

	err := p.redisClient.AppendToNotifyWorkQueue(ctx, []byte("new:"+incidentID))
	if err != nil {
		logger.Error(err, "error appending new incident to notify work queue", "incidentID", incidentID)
		return
	}
}

func (p *PodEventProcessor) NotifyResolvedIncident(ctx context.Context, incidentID string) {
	logger := log.Log.WithName("event-processor")

	err := p.redisClient.AppendToNotifyWorkQueue(ctx, []byte("resolved:"+incidentID))
	if err != nil {
		logger.Error(err, "error appending resolved incident to notify work queue", "incidentID", incidentID)
		return
	}
}
