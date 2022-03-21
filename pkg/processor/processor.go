package processor

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

type EnqueueDetailOptions struct {
	ContainerNamesToFetchLogs []string
}

type Interface interface {
	NotifyNewIncident(context.Context, string)

	NotifyResolvedIncident(context.Context, string)
}

func (e *EnqueueDetailOptions) SetContainerName(podOptions *corev1.PodLogOptions) {
	if len(e.ContainerNamesToFetchLogs) > 0 {
		podOptions.Container = e.ContainerNamesToFetchLogs[0]
	}
}
