package models

import (
	"fmt"
	"strings"
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
		lastSeen = *i.Events[0].LastSeen

		for _, e := range i.Events {
			if e.LastSeen.After(lastSeen) {
				lastSeen = *e.LastSeen
			}
		}
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
		Summary:                 i.toExternalSummary(),
	}
}

func (i *Incident) ToAPIType() *types.Incident {
	incident := &types.Incident{
		IncidentMeta: i.ToAPITypeMeta(),
		Pods:         i.getUniquePods(),
	}

	incident.Detail = "The release failed"

	for _, e := range i.Events {
		if e.IsPrimaryCause {
			incident.Detail = e.Detail
			break
		}
	}

	return incident
}

func (i *Incident) GetInternalSummary() string {
	summary := "The release failed"

	for _, e := range i.Events {
		if e.IsPrimaryCause {
			summary = e.Summary
			break
		}
	}

	return summary
}

func (i *Incident) getUniquePods() []string {
	uniquePods := make(map[string]string, 0)

	for _, ev := range i.Events {
		uniquePods[ev.PodName] = ev.PodName
	}

	res := make([]string, 0)

	for _, podName := range uniquePods {
		res = append(res, podName)
	}

	return res
}

func (i *Incident) toExternalSummary() string {
	uniquePods := i.getUniquePods()

	// if the incident is part of a deployment, we count the number of unique
	// pods involved and generate a message.
	if strings.ToLower(string(i.InvolvedObjectKind)) == "deployment" {
		if len(uniquePods) > 1 {
			if i.Severity == types.SeverityCritical {
				return fmt.Sprintf(
					"Your application %s in namespace %s is currently experiencing downtime. %d replicas are crashing due to: %s",
					i.ReleaseName, i.ReleaseNamespace, len(uniquePods), i.GetInternalSummary(),
				)
			} else {
				return fmt.Sprintf(
					"%d replicas for the application %s in namespace %s have crashed due to: %s",
					len(uniquePods), i.ReleaseName, i.ReleaseNamespace, i.GetInternalSummary(),
				)
			}
		} else {
			if i.Severity == types.SeverityCritical {
				return fmt.Sprintf(
					"Your application %s in namespace %s is currently experiencing downtime. The application crashed due to: %s",
					i.ReleaseName, i.ReleaseNamespace, i.GetInternalSummary(),
				)
			} else {
				return fmt.Sprintf(
					"Your application %s in namespace %s has crashed due to: %s",
					i.ReleaseName, i.ReleaseNamespace, i.GetInternalSummary(),
				)
			}
		}
	}

	// if the incident is part of a job, we indicate that this was part of a job run
	if strings.ToLower(string(i.InvolvedObjectKind)) == "job" {
		return fmt.Sprintf(
			"A job run for %s in namespace %s crashed due to: %s",
			i.ReleaseName, i.ReleaseNamespace, i.GetInternalSummary(),
		)
	}

	// otherwise, we just incidate that a single replica failed
	return fmt.Sprintf(
		"Your application %s in namespace %s has crashed due to: %s",
		i.ReleaseName, i.ReleaseNamespace, i.GetInternalSummary(),
	)
}
