package incident

import (
	"fmt"
	"strings"
	"time"

	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/repository"
	"github.com/porter-dev/porter-agent/internal/utils"
	"github.com/porter-dev/porter-agent/pkg/alerter"
	"github.com/porter-dev/porter-agent/pkg/event"
	"k8s.io/client-go/kubernetes"
)

type IncidentDetector struct {
	KubeClient  *kubernetes.Clientset
	KubeVersion KubernetesVersion
	EventStore  event.EventStore
	Repository  *repository.Repository
	Alerter     *alerter.Alerter
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

		alertedEvents = append(alertedEvents, e)
	}

	if len(alertedEvents) == 0 {
		return nil
	}

	// at this point, populate the owner reference for the first alerted event - we assume that
	// all alerted events have the same owner
	err := alertedEvents[0].Populate(*d.KubeClient)

	if err != nil {
		return nil
	}

	// populate all other alerted events with the same data
	for i := range alertedEvents {
		alertedEvents[i].Pod = alertedEvents[0].Pod
		alertedEvents[i].Owner = alertedEvents[0].Owner
		alertedEvents[i].ReleaseName = alertedEvents[0].ReleaseName
		alertedEvents[i].ChartName = alertedEvents[0].ChartName
		alertedEvents[i].ChartVersion = alertedEvents[0].ChartVersion
	}

	// get event matches
	matches := make(map[event.FilteredEvent]*EventMatch)

	for _, e := range alertedEvents {
		matchCandidate := GetEventMatchFromEvent(d.KubeVersion, d.KubeClient, e)

		// we only add match candidates which have a primary cause at the moment
		if matchCandidate != nil && matchCandidate.IsPrimaryCause {
			matches[*e] = matchCandidate
		}
	}

	fmt.Println("LENGTH OF MATCHES IS", len(matches))

	// iterate through incident events
	for alertedEvent, match := range matches {
		// construct the basic incident event model
		incident := getIncidentMetaFromEvent(&alertedEvent)
		incident.Events = matchesToIncidentEvent(d.KubeVersion, map[event.FilteredEvent]*EventMatch{
			alertedEvent: match,
		})

		ownerRef := alertedEvent.Owner

		switch strings.ToLower(ownerRef.Kind) {
		case "deployment":
			fmt.Printf("determing if deployment %s is failing\n", ownerRef.Name)

			// if the deployment is in a failure state, create a high severity incident
			if isDeploymentFailing(d.KubeClient, ownerRef.Namespace, ownerRef.Name) {
				incident.Severity = models.SeverityCritical
				incident.InvolvedObjectKind = models.InvolvedObjectDeployment
				incident.InvolvedObjectName = ownerRef.Name
				incident.InvolvedObjectNamespace = ownerRef.Namespace

				err := d.saveIncident(incident, ownerRef)

				if err != nil {
					return err
				}

				continue
			}
		case "job":
			incident.Severity = models.SeverityNormal
			incident.InvolvedObjectKind = models.InvolvedObjectJob
			incident.InvolvedObjectName = ownerRef.Name
			incident.InvolvedObjectNamespace = ownerRef.Namespace

			err := d.saveIncident(incident, ownerRef)

			if err != nil {
				return err
			}

			continue
		}

		// if the controller cases did not match, we simply store a pod-based incident
		incident.Severity = models.SeverityNormal
		incident.InvolvedObjectKind = models.InvolvedObjectPod
		incident.InvolvedObjectName = alertedEvent.PodName
		incident.InvolvedObjectNamespace = alertedEvent.PodNamespace

		err := d.saveIncident(incident, ownerRef)

		if err != nil {
			return err
		}
	}

	return nil
}

func (d *IncidentDetector) saveIncident(incident *models.Incident, ownerRef *event.EventOwner) error {
	// if mergeWithMatchingIncident returns a non-nil incident, then we simply update the incident in the DB
	if mergedIncident := d.mergeWithMatchingIncident(incident, ownerRef); mergedIncident != nil {
		matchedIncidentID := mergedIncident.ID
		incident, err := d.Repository.Incident.UpdateIncident(mergedIncident)

		if err != nil {
			return err
		}

		fmt.Println("INCIDENTS MATCHED:", matchedIncidentID, incident.ID)

		return d.Alerter.HandleIncident(incident)
	}

	fmt.Println("CREATING NEW INCIDENT")

	incident, err := d.Repository.Incident.CreateIncident(incident)

	if err != nil {
		return err
	}

	return d.Alerter.HandleIncident(incident)
}

func (d *IncidentDetector) mergeWithMatchingIncident(incident *models.Incident, ownerRef *event.EventOwner) *models.Incident {
	// we look for a matching incident - the matching incident should refer to the same
	// release name and namespace, should be active, and the incident event should have
	// a primary cause event with the same summary as the candidate incident.
	statusActive := models.IncidentStatusActive

	candidateMatches, err := d.Repository.Incident.ListIncidents(&utils.ListIncidentsFilter{
		Status:           &statusActive,
		ReleaseName:      &incident.ReleaseName,
		ReleaseNamespace: &incident.ReleaseNamespace,
	})

	if err != nil {
		return nil
	}

	var primaryCauseSummary string

	for _, currIncidentEvent := range incident.Events {
		if currIncidentEvent.IsPrimaryCause {
			primaryCauseSummary = currIncidentEvent.Summary
			break
		}
	}

	for _, candidateMatch := range candidateMatches {
		for _, candidateMatchEvent := range candidateMatch.Events {
			if candidateMatchEvent.IsPrimaryCause && candidateMatchEvent.Summary == primaryCauseSummary {
				fmt.Println("GOT MATCHING INCIDENT", candidateMatch)

				// in this case, we've found a match, and we merge and return
				candidateMatch.LastSeen = incident.LastSeen
				mergedEvents := mergeEvents(candidateMatch.Events, incident.Events)
				candidateMatch.Events = mergedEvents

				// if there are different pods listed in the events, we promote this to a "Deployment" event
				if numDistinctPods(mergedEvents) > 1 {
					candidateMatch.InvolvedObjectKind = models.InvolvedObjectKind(ownerRef.Kind)
					candidateMatch.InvolvedObjectName = ownerRef.Name
					candidateMatch.InvolvedObjectNamespace = ownerRef.Namespace
				}

				return candidateMatch
			}
		}
	}

	return nil
}

func mergeEvents(events1, events2 []models.IncidentEvent) []models.IncidentEvent {
	// we construct a key for events1 by looking at the pod name, namespace, primary cause, and
	// summary fields
	keyMap := make(map[string]models.IncidentEvent)

	for _, e1 := range events1 {
		keyMap[fmt.Sprintf("%s/%s-%v-%s", e1.PodName, e1.PodNamespace, e1.IsPrimaryCause, e1.Summary)] = e1
	}

	// any matched events are updated, other events are appended
	now := time.Now()

	for _, e2 := range events2 {
		key := fmt.Sprintf("%s/%s-%v-%s", e2.PodName, e2.PodNamespace, e2.IsPrimaryCause, e2.Summary)

		if e1, exists := keyMap[key]; exists {
			e1.LastSeen = &now
			e1.Detail = e2.Detail
		} else {
			keyMap[key] = e2
		}
	}

	res := make([]models.IncidentEvent, 0)

	for _, e := range keyMap {
		res = append(res, e)
	}

	return res
}

func numDistinctPods(events []models.IncidentEvent) int {
	podMap := make(map[string]string)

	for _, e := range events {
		key := fmt.Sprintf("%s/%s", e.PodNamespace, e.PodName)
		podMap[key] = key
	}

	return len(podMap)
}
