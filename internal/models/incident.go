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
	lastSeen := time.Now()

	if len(i.Events) > 0 {
		// TODO: get the most recent event, not just the first
		lastSeen = *i.Events[0].LastSeen
	}

	// TODO: generate a better summary
	summary := "The release failed"

	if len(i.Events) > 0 {
		summary = i.Events[0].Summary
	}

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
		LastSeen:                &lastSeen,
		Summary:                 summary,
	}
}

func (i *Incident) ToAPIType() *types.Incident {
	involvedPods := make(map[string]string)

	for _, event := range i.Events {
		involvedPods[event.PodName] = event.PodName
	}

	pods := make([]string, 0)

	for podName := range involvedPods {
		pods = append(pods, podName)
	}

	// TODO: generate better details
	detail := "The release failed"

	if len(i.Events) > 0 {
		detail = i.Events[0].Summary
	}

	return &types.Incident{
		IncidentMeta: i.ToAPITypeMeta(),
		Pods:         pods,
		Detail:       detail,
	}
}
