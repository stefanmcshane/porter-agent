package models

type EventResourceType string

func (e *EventResourceType) String() string {
	return string(*e)
}

const (
	PodResource  EventResourceType = "Pod"
	HAPResource  EventResourceType = "HPA"
	NodeResource EventResourceType = "Node"

	EventCritical EventCriticality = "Critical"
	EventNormal   EventCriticality = "Normal"

	UnhealthyToHealthyTransitionMessage string = "Pod transitioned from unhealthy to healthy state"
)

type EventCriticality string

func (e *EventCriticality) String() string {
	return string(*e)
}

type EventDetails struct {
	ResourceType EventResourceType `json:"resource_type"`
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Cluster      string            `json:"cluster"`
	Message      string            `json:"message"`
	Reason       string            `json:"reason"`
	Data         []string          `json:"data"`
	Critical     bool              `json:"critical"`
	Timestamp    string            `json:"timestamp"`
	EventType    EventCriticality  `json:"event_type"`
	Phase        string            `json:"pod_phase"`
	Status       string            `json:"pod_status"`
}
