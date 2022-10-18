package types

import "time"

type EventType string

const (
	EventTypeIncident           EventType = "incident"
	EventTypeIncidentResolved   EventType = "incident_resolved"
	EventTypeDeploymentStarted  EventType = "deployment_started"
	EventTypeDeploymentFinished EventType = "deployment_finished"
	EventTypeDeploymentErrored  EventType = "deployment_errored"
)

type Event struct {
	Type             EventType              `json:"type"`
	Version          string                 `json:"version"`
	ReleaseName      string                 `json:"release_name"`
	ReleaseNamespace string                 `json:"release_namespace"`
	Timestamp        *time.Time             `json:"timestamp"`
	Data             map[string]interface{} `json:"data"`
}

type ListEventsRequest struct {
	*PaginationRequest
	ReleaseName      *string `schema:"release_name"`
	ReleaseNamespace *string `schema:"release_namespace"`
	Type             *string `schema:"type"`
}

type ListEventsResponse struct {
	Events     []*Event            `json:"events" form:"required"`
	Pagination *PaginationResponse `json:"pagination"`
}
