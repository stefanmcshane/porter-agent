package processor

import (
	"context"

	"github.com/porter-dev/porter-agent/pkg/models"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type EnqueueDetailOptions struct {
	ContainerNamesToFetchLogs []string
}

type Interface interface {
	// Used in case of normal events to store and update logs
	EnqueueDetails(context context.Context, object types.NamespacedName, options *EnqueueDetailOptions)
	// to trigger actual request for porter server in case of
	// a Delete or Failed/Unknown Phase
	AddToWorkQueue(context context.Context, object types.NamespacedName, details models.EventDetails)
}

func (e *EnqueueDetailOptions) SetContainerName(podOptions *corev1.PodLogOptions) {
	if len(e.ContainerNamesToFetchLogs) > 0 {
		podOptions.Container = e.ContainerNamesToFetchLogs[0]
	}
}
