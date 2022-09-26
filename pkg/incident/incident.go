package incident

import (
	"context"

	"github.com/porter-dev/porter-agent/pkg/event"
	"k8s.io/client-go/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
)

type IncidentSeverity string

const (
	IncidentSeverityCritical IncidentSeverity = "critical"
	IncidentSeverityWarning  IncidentSeverity = "warning"
)

// TODO: create incident API object
type Incident struct {
	Message string
	Reason  string

	Severity IncidentSeverity
}

type IncidentDetector struct {
	KubeClient *kubernetes.Clientset
	EventStore event.EventStore
}

// DetectDeploymentIncident returns an incident if one should be triggered, if there is no incident it will return
// a nil incident and nil error message.
//
// It determines if an incident should be alerted based on the following algorithm:
// 1. What is the event type?
//     1. `Normal`: do not alert
//     2. `Critical`: 2
// 2. Did the event trigger a container restart or prevent the pod from starting up?
//     1. Yes: 3
//     2. No: do not alert
// 3. Are there more pods unavailable than the deployments `maxUnavailable` field permits?
//     1. Yes: 4
//     2. No: 5
// 4. Trigger an immediate alert and create a critical incident for the user.
// 5. Query for past events from this pod. If the event has been triggered a certain number of times
//    (configurable) in a certain time window (configurable), create a warning incident for the user.
func (d *IncidentDetector) DetectDeploymentIncident(e *event.Event, o *event.EventOwner) (*Incident, error) {
	// if the event severity is low, do not alert
	if e.Severity == event.EventSeverityLow {
		return nil, nil
	}

	// if the event neither triggered a container restart or prevented the pod from starting up,
	// do not alert
	if !d.didPreventStartup(e) && !d.didTriggerRestart(e) {
		return nil, nil
	}

	// if the deployment is in a failure state, create a high severity incident
	if d.isDeploymentFailing(o) {
		// TODO: compute better error message
		return &Incident{
			Message:  "The deployment is failing!",
			Reason:   "Failing",
			Severity: IncidentSeverityCritical,
		}, nil
	}

	// otherwise query for past events, and determine if this should be alerted
	// TODO: implement
	return nil, nil
}

func (d *IncidentDetector) didPreventStartup(e *event.Event) bool {
	// TODO: implement
	return false
}

func (d *IncidentDetector) didTriggerRestart(e *event.Event) bool {
	// TODO: implement
	return false
}

// we define a deployment as "failing" if it has less than maxUnavailable replicas which
// are reporting a not ready status
func (d *IncidentDetector) isDeploymentFailing(o *event.EventOwner) bool {
	// query the deployment from the live cluster state
	depl, err := d.KubeClient.AppsV1().Deployments(o.Namespace).Get(
		context.Background(),
		o.Name,
		v1.GetOptions{},
	)

	if err != nil {
		// TODO: this case should trigger a warning, as it indicates an invalid configuration for
		// the agent
		return false
	}

	// determine if the deployment has an appropriate number of ready replicas
	minUnavailable := *(depl.Spec.Replicas) - getMaxUnavailable(depl)

	return minUnavailable <= depl.Status.ReadyReplicas
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
