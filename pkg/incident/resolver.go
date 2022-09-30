package incident

import (
	"fmt"
	"time"

	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/repository"
	"github.com/porter-dev/porter-agent/internal/utils"
	"github.com/porter-dev/porter-agent/pkg/alerter"
	"k8s.io/client-go/kubernetes"
)

type IncidentResolver struct {
	KubeClient  *kubernetes.Clientset
	KubeVersion KubernetesVersion
	Repository  *repository.Repository
	Alerter     *alerter.Alerter
}

// INCIDENT_REPEAT_BUFFER_HOURS are the number of hours we provide as buffer to determine whether the issue
// has appeared again for a specific pod. This is used for "Warning" incidents that the user may not immediately
// observe, and so we allow for a healthy buffer before resolving this incident.
//
// For example, if Pod A experiences an OOMKilled error (but the deployment is otherwise healthy), we do not
// immediately resolve the incident once the pod is removed or restarts in a healthy state. We will wait for 24 hours
// until we see this incident occur again.
//
// This value may be made configurable in the future for different alerting configurations. We make this 23 hours
// as email digests for warning incidents may be sent out daily.
const INCIDENT_REPEAT_BUFFER_HOURS = 23

// CRITICAL_BUFFER_MINUTES are the number of minutes we provide as grace period to check whether or not a
// critical incident has been resolved
const CRITICAL_BUFFER_MINUTES = 6

func (r *IncidentResolver) Run() error {
	// get all active incidents
	// TODO: pagination

	statusActive := models.IncidentStatusActive

	activeIncidents, err := r.Repository.Incident.ListIncidents(&utils.ListIncidentsFilter{
		Status: &statusActive,
	})

	if err != nil {
		return err
	}

	for _, activeIncident := range activeIncidents {
		fmt.Printf("checking whether incident %s is resolved\n", activeIncident.UniqueID)

		if r.isResolved(activeIncident) {
			fmt.Println("the incident was resolved!")

			if err := r.handleResolved(activeIncident); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *IncidentResolver) handleResolved(incident *models.Incident) error {
	resolvedTime := time.Now()
	incident.ResolvedTime = &resolvedTime
	incident.IncidentStatus = models.IncidentStatusResolved

	_, err := r.Repository.Incident.UpdateIncident(incident)

	if err != nil {
		return err
	}

	return r.Alerter.HandleResolved(incident)
}

func (r *IncidentResolver) isResolved(incident *models.Incident) bool {
	// switch on the incident type
	switch incident.InvolvedObjectKind {
	case models.InvolvedObjectDeployment:
		return r.isDeploymentResolved(incident)
	case models.InvolvedObjectJob:
		return r.isJobResolved(incident)
	case models.InvolvedObjectPod:
		return r.isPodResolved(incident)
	}

	return false
}

func (r *IncidentResolver) isDeploymentResolved(incident *models.Incident) bool {
	// if this is a critical incident, we check whether the deployment has been running
	// successfully for at least the critical buffer window
	if incident.Severity == models.SeverityCritical {
		return !r.isWithinCriticalBufferWindow(incident.LastSeen) &&
			!isDeploymentFailing(r.KubeClient, incident.InvolvedObjectNamespace, incident.InvolvedObjectName)
	}

	// If this is not a critical incident, we check the buffer window for when this was last seen, because pods will
	// continue to trigger
	return !r.isWithinBufferWindow(incident.LastSeen)
}

// TODO: the casing for jobs should involve alerting when a certain number of job runs have triggered a
// failure, which should be a configurable parameter. Right now we simply case on the pod buffer.
func (r *IncidentResolver) isJobResolved(incident *models.Incident) bool {
	return r.isPodResolved(incident)
}

func (r *IncidentResolver) isPodResolved(incident *models.Incident) bool {
	// All we have to check is whether or not the last time this incident was seen is within the buffer window,
	// because if the pods continue to fail for the same reasons the incident will be updated with a new LastSeen
	// time. This also applies to the case where pods have since been deleted.
	return !r.isWithinBufferWindow(incident.LastSeen)
}

func (r *IncidentResolver) isWithinBufferWindow(lastSeen *time.Time) bool {
	return lastSeen.Add(INCIDENT_REPEAT_BUFFER_HOURS * time.Hour).After(time.Now())
}

func (r *IncidentResolver) isWithinCriticalBufferWindow(lastSeen *time.Time) bool {
	return lastSeen.Add(CRITICAL_BUFFER_MINUTES * time.Minute).After(time.Now())
}
