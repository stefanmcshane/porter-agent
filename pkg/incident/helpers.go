package incident

import (
	"context"
	"fmt"
	"time"

	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/pkg/event"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// we define a deployment as "failing" if it has less than maxUnavailable replicas which
// are reporting a not ready status
func isDeploymentFailing(kubeClient *kubernetes.Clientset, deplNamespace, deplName string) bool {
	// query the deployment from the live cluster state
	depl, err := kubeClient.AppsV1().Deployments(deplNamespace).Get(
		context.Background(),
		deplName,
		v1.GetOptions{},
	)

	if err != nil {
		// TODO: this case should trigger a warning, as it indicates an invalid configuration for
		// the agent
		return false
	}

	// determine if the deployment has an appropriate number of ready replicas
	minUnavailable := *(depl.Spec.Replicas) - getMaxUnavailable(depl)

	fmt.Printf("min unavailable is %d, ready replicas are %d\n", minUnavailable, depl.Status.ReadyReplicas)

	return depl.Status.ReadyReplicas < minUnavailable
}

func getMaxUnavailable(deployment *appsv1.Deployment) int32 {
	if deployment.Spec.Strategy.Type != appsv1.RollingUpdateDeploymentStrategyType || *(deployment.Spec.Replicas) == 0 {
		return int32(0)
	}

	desired := *(deployment.Spec.Replicas)
	maxUnavailable := deployment.Spec.Strategy.RollingUpdate.MaxUnavailable

	unavailable, err := intstrutil.GetScaledValueFromIntOrPercent(intstrutil.ValueOrDefault(maxUnavailable, intstrutil.FromInt(0)), int(desired), false)

	if err != nil {
		return 0
	}

	return int32(unavailable)
}

func matchesToIncidentEvent(k8sVersion KubernetesVersion, es map[event.FilteredEvent]*EventMatch) []models.IncidentEvent {
	res := make([]models.IncidentEvent, 0)

	for filteredEvent, match := range es {
		res = append(res, models.IncidentEvent{
			Summary:        match.Summary,
			Detail:         match.DetailGenerator(&filteredEvent),
			PodName:        filteredEvent.PodName,
			PodNamespace:   filteredEvent.PodNamespace,
			IsPrimaryCause: match.IsPrimaryCause,
		})
	}

	return res
}

func getIncidentMetaFromEvent(e *event.FilteredEvent) *models.Incident {
	res := models.NewIncident()

	res.IncidentStatus = models.IncidentStatusActive
	res.ReleaseName = e.ReleaseName
	res.ReleaseNamespace = e.Owner.Namespace
	res.ChartName = e.ChartName
	res.Severity = models.SeverityType(e.Severity)

	lastSeen := time.Now()

	res.LastSeen = &lastSeen

	return res
}
