package models

import (
	"time"

	"gorm.io/gorm"
)

type SeverityType string

const (
	SeverityCritical SeverityType = "critical"
	SeverityNormal   SeverityType = "normal"
)

type InvolvedObjectKind string

const (
	InvolvedObjectDeployment InvolvedObjectKind = "deployment"
	InvolvedObjectJob        InvolvedObjectKind = "job"
	InvolvedObjectPod        InvolvedObjectKind = "pod"
)

type IncidentStatus string

const (
	IncidentStatusResolved IncidentStatus = "resolved"
	IncidentStatusActive   IncidentStatus = "active"
)

type Incident struct {
	gorm.Model

	UniqueID string `gorm:"unique"`

	LastAlerted *time.Time
	LastSeen    *time.Time

	ResolvedTime *time.Time

	IncidentStatus IncidentStatus

	ReleaseName      string
	ReleaseNamespace string
	ChartName        string

	InvolvedObjectKind      InvolvedObjectKind
	InvolvedObjectName      string
	InvolvedObjectNamespace string

	Severity SeverityType

	Events []IncidentEvent
}

func NewIncident() *Incident {
	randStr, _ := GenerateRandomBytes(16)

	return &Incident{
		UniqueID: randStr,
	}
}
