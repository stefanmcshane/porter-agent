package models

import (
	"time"

	"github.com/porter-dev/porter-agent/api/server/types"
	"gorm.io/gorm"
)

type Incident struct {
	gorm.Model

	UniqueID string `gorm:"unique"`

	LastAlerted *time.Time
	LastSeen    *time.Time

	ResolvedTime *time.Time

	IncidentStatus types.IncidentStatus

	ReleaseName      string
	ReleaseNamespace string
	ChartName        string

	InvolvedObjectKind      types.InvolvedObjectKind
	InvolvedObjectName      string
	InvolvedObjectNamespace string

	Severity types.SeverityType

	Events []IncidentEvent
}

func NewIncident() *Incident {
	randStr, _ := GenerateRandomBytes(16)

	return &Incident{
		UniqueID: randStr,
	}
}

func (i *Incident) ToAPITypeMeta() *types.IncidentMeta {
	return &types.IncidentMeta{
		ID:                      i.UniqueID,
		ReleaseName:             i.ReleaseName,
		ReleaseNamespace:        i.ReleaseNamespace,
		UpdatedAt:               i.UpdatedAt,
		CreatedAt:               i.CreatedAt,
		ChartName:               i.ChartName,
		Status:                  i.IncidentStatus,
		InvolvedObjectKind:      i.InvolvedObjectKind,
		InvolvedObjectName:      i.InvolvedObjectName,
		InvolvedObjectNamespace: i.InvolvedObjectNamespace,
		Severity:                i.Severity,

		// LastSeen: ,
		// Status: ,
		// Summary: ,
	}
}

func (i *Incident) ToAPIType() *types.Incident {
	incident := &types.Incident{
		IncidentMeta: i.ToAPITypeMeta(),
	}

	if len(i.Events) > 0 {
		incident.Detail = i.Events[0].Detail
	}

	for _, ev := range i.Events {
		incident.Pods = append(incident.Pods, ev.PodName)
	}

	return incident
}
