package processor

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/porter-dev/porter-agent/pkg/redis"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PodEventProcessor struct {
	kubeClient   *kubernetes.Clientset
	redisClient  *redis.Client
	resourceType string
}

func NewPodEventProcessor(config *rest.Config) Interface {
	return &PodEventProcessor{
		kubeClient:   kubernetes.NewForConfigOrDie(config),
		resourceType: "pod",
		redisClient:  redis.NewClient("127.0.0.1", "6379", "", "", redis.PODSTORE, int64(100)),
	}
}

// Used in case of normal events to store and update logs
func (p *PodEventProcessor) EnqueueWithLogLines(ctx context.Context, object types.NamespacedName) {
	maxTailLines := new(int64)
	*maxTailLines = 100
	logger := log.FromContext(ctx)

	req := p.kubeClient.
		CoreV1().
		Pods(object.Namespace).
		GetLogs(object.Name, &corev1.PodLogOptions{
			TailLines: maxTailLines,
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
	err = p.redisClient.AppendAndTrimDetails(ctx, p.resourceType, object.Namespace, object.Name, strings.Split(strLogs, "\n"))
	if err != nil {
		logger.Error(err, "unable to append logs to the store")
		return
	}
}

// to trigger actual request for porter server in case of
// a Delete or Failed/Unknown Phase
func (p *PodEventProcessor) TriggerNotifyForFatalEvent(ctx context.Context, object types.NamespacedName, details map[string]interface{}) {
	logger := log.FromContext(ctx)
	logger.Info("notification triggered")

	logger.Info("current pod condition", "details", details)
	// TODO: Implement
}
