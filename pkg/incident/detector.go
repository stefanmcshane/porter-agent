package incident

import (
	"context"
	"fmt"
	"strings"

	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/repository"
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

type IncidentDetector struct {
	KubeClient  *kubernetes.Clientset
	KubeVersion KubernetesVersion
	EventStore  event.EventStore
	Repository  repository.Repository
}

// DetectIncident returns an incident if one should be triggered, if there is no incident it will return
// a nil incident and nil error message.
//
// It determines if an incident should be alerted based on the following algorithm:
// 1. What is the event type?
//     1. `Normal`: do not alert
//     2. `Critical`: 2
// 2. Did the event trigger a container restart or prevent the pod from starting up?
//     1. Yes: 3
//     2. No: do not alert
// 3a. (If Deployment) Are there more pods unavailable than the deployments `maxUnavailable` field permits?
//     1. Yes: 4
//     2. No: 5
// 3b. (If Job) Does the alerting threshold match configuration for this job?
//     1. Yes: 4
//     2. No: 5
// 4. Trigger an immediate alert and create a critical incident for the user.
// 5. Query for past events from this pod. If the event has been triggered a certain number of times
//    (configurable) in a certain time window (configurable), create a warning incident for the user.
func (d *IncidentDetector) DetectIncident(es []*event.FilteredEvent) error {
	alertedEvents := make([]*event.FilteredEvent, 0)

	for _, e := range es {
		fmt.Println("processing:", e.KubernetesReason, e.KubernetesMessage)

		// if the event severity is low, do not alert
		if e.Severity == event.EventSeverityLow {
			continue
		}

		// if the event neither triggered a container restart or prevented the pod from starting up,
		// do not alert
		// if !d.didPreventStartup(e) && !d.didTriggerRestart(e) {
		// 	continue
		// }

		alertedEvents = append(alertedEvents, e)
	}

	if len(alertedEvents) == 0 {
		return nil
	}

	// at this point, populate the owner reference for the first alerted event - we assume that
	// all alerted events have the same owner
	err := alertedEvents[0].PopulateEventOwner(*d.KubeClient)

	if err != nil {
		return nil
	}

	// get event matches
	matches := make(map[event.FilteredEvent]*EventMatch)

	for _, e := range alertedEvents {
		matchCandidate := GetEventMatchFromEvent(d.KubeVersion, e)

		if matchCandidate != nil {
			matches[*e] = matchCandidate
		}
	}

	// if the length of matches is not 0, we have a matched incident
	if len(matches) != 0 {
		switch strings.ToLower(alertedEvents[0].Owner.Kind) {
		case "deployment":
			fmt.Printf("determing if deployment %s is failing\n", alertedEvents[0].Owner.Name)

			// if the deployment is in a failure state, create a high severity incident
			if d.isDeploymentFailing(alertedEvents[0].Owner) {
				incident := &models.Incident{
					// Message:        "The deployment is failing!",
					// Reason:         "Failing",
					// Severity:       IncidentSeverityCritical,
					// IncidentEvents: matchesToIncidentEvent(d.KubeVersion, matches),
				}

				return d.saveIncident(incident)
			}
		case "job":
			// if the owner is a job, we always create a job-based incident as they typically run
			// only once on Porter
			incident := &models.Incident{
				// Message:        "The deployment is failing!",
				// Reason:         "Failing",
				// Severity:       IncidentSeverityCritical,
				// IncidentEvents: matchesToIncidentEvent(d.KubeVersion, matches),
			}

			return d.saveIncident(incident)
		}

		// if the controller cases did not match, we simply store a pod-based incident
		incident := &models.Incident{
			// Message:        "The deployment is failing!",
			// Reason:         "Failing",
			// Severity:       IncidentSeverityCritical,
			// IncidentEvents: matchesToIncidentEvent(d.KubeVersion, matches),
		}

		return d.saveIncident(incident)
	}

	return nil
}

func (d *IncidentDetector) didPreventStartup(e *event.FilteredEvent) bool {
	// TODO: implement
	return false
}

func (d *IncidentDetector) didTriggerRestart(e *event.FilteredEvent) bool {
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
			Summary:      match.Summary,
			Detail:       match.DetailGenerator(&filteredEvent),
			PodName:      filteredEvent.PodName,
			PodNamespace: filteredEvent.PodNamespace,
		})
	}

	return res
}

func (d *IncidentDetector) saveIncident(incident *models.Incident) error {
	// if mergeWithMatchingIncident returns a non-nil incident, then we simply update the incident in the DB
	if mergedIncident := d.mergeWithMatchingIncident(incident); mergedIncident != nil {
		// TODO: switch this to update incident
		_, err := d.Repository.Incident.UpdateIncident(mergedIncident)

		return err
	}

	_, err := d.Repository.Incident.CreateIncident(incident)

	return err
}

func (d *IncidentDetector) mergeWithMatchingIncident(incident *models.Incident) *models.Incident {
	// we look for a matching incident - the matching incident should refer to the same
	// release name and namespace, should be active, and the incident event should have
	// a primary cause event with the same summary as the candidate incident.

	// if a matching incident is found and currently refers to a pod, and the candidate incident
	// refers to a different pod, this is promoted to a "deployment" incident.

	return nil
}
