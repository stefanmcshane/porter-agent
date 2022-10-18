package models

import (
	"time"

	"gorm.io/gorm"
)

type EventType string

const (
	EventTypeIncident           EventType = "incident"
	EventTypeIncidentResolved   EventType = "incident_resolved"
	EventTypeDeploymentStarted  EventType = "deployment_started"
	EventTypeDeploymentFinished EventType = "deployment_finished"
	EventTypeDeploymentErrored  EventType = "deployment_errored"
)

type Event struct {
	gorm.Model

	UniqueID string `gorm:"unique"`

	Version string
	Type    EventType

	ReleaseName      string
	ReleaseNamespace string
	Timestamp        *time.Time
	Description      string
	HasLogs          bool

	Data []byte
}
