package processor

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Interface interface {
	// Used in case of normal events to store and update logs
	EnqueueWithLogLines(context context.Context, object types.NamespacedName)
	// to trigger actual request for porter server in case of
	// a Delete or Failed/Unknown Phase
	TriggerNotifyForFatalEvent(object types.NamespacedName, details map[string]interface{})
}

type PodEventProcessor struct {
	client *kubernetes.Clientset
}

func NewPodEventProcessor(config *rest.Config) Interface {
	return &PodEventProcessor{
		client: kubernetes.NewForConfigOrDie(config),
	}
}

// Used in case of normal events to store and update logs
func (p *PodEventProcessor) EnqueueWithLogLines(ctx context.Context, object types.NamespacedName) {
	maxTailLines := new(int64)
	*maxTailLines = 100
	logger := log.Log.WithValues("pod processor")

	req := p.client.
		CoreV1().
		Pods(object.Namespace).
		GetLogs(object.Name, &corev1.PodLogOptions{
			TailLines: maxTailLines,
		})

	podLogs, err := req.Stream(ctx)
	if err != nil {
		logger.Error(err, "error streaming logs")
	}
	defer podLogs.Close()

	var logs []byte
	_, err = podLogs.Read(logs)
	if err != nil {
		logger.Error(err, "unable to read logs")
	}

	logger.Info("Successfully fetched logs")
	// update logs in the redis store
}

// to trigger actual request for porter server in case of
// a Delete or Failed/Unknown Phase
func (p *PodEventProcessor) TriggerNotifyForFatalEvent(object types.NamespacedName, details map[string]interface{}) {
	logger := log.Log.WithValues("pod notify trigger")
	logger.Info("notification triggered")

	logger.Info("current pod condition", "details", details)
	panic("not implemented") // TODO: Implement
}
