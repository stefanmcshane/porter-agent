package event

import (
	"time"

	"k8s.io/client-go/kubernetes"
)

type EventSeverity string

const (
	EventSeverityCritical EventSeverity = "critical"
	EventSeverityHigh     EventSeverity = "high"
	EventSeverityLow      EventSeverity = "low"
)

type Event struct {
	PodName      string
	PodNamespace string

	Reason  string
	Message string

	Severity EventSeverity

	Timestamp *time.Time
}

type EventOwner struct {
	Namespace, Name, Kind string
}

func (e *Event) GetEventOwner(k8sClient kubernetes.Clientset) (*EventOwner, error) {
	return nil, nil
}

type EventStore interface {
	Store(e *Event) error
	GetEventsByPodName(namespace, name string) *Event
	GetEventsByOwner(owner *EventOwner) *Event
}
