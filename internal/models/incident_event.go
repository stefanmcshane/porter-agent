package models

import (
	"time"

	"github.com/porter-dev/porter-agent/api/server/types"
	"gorm.io/gorm"
)

type IncidentEvent struct {
	gorm.Model

	UniqueID   string `gorm:"unique"`
	IncidentID uint

	LastSeen *time.Time

	// Each incident event corresponds to a single pod name and namespace.
	PodName      string
	PodNamespace string

	// Summary is a high-level reason for the incident. Each incident event that is designated
	// as a "primary cause" should have the same summary across multiple pods.
	Summary string

	// Detail contains more information about the incident - for example, which exit code the pod
	// exited with, how long the pod has been stuck in pending, etc.
	Detail string

	// IsPrimaryCause informs whether this event was the primary cause of the incident,
	// or this is an auxiliary event. For example, in some cases we may process that a
	// pod is in a backoff state before determining the cause of that backoff state - in
	// those cases, an incident will have multiple incident events, but each incident should
	// only have one primary cause per pod.
	//
	// When taken together, the pod owner, summary, and primary cause field determine whether two events should be
	// considered part of the same incident, or considered two different incidents.
	IsPrimaryCause bool
}

func (e *IncidentEvent) ToAPIType() *types.IncidentEvent {
	return &types.IncidentEvent{
		LastSeen:     e.LastSeen,
		PodName:      e.PodName,
		PodNamespace: e.PodNamespace,
		Summary:      e.Summary,
		Detail:       e.Detail,
	}
}
