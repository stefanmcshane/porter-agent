package models

type EventResourceType string

func (e *EventResourceType) String() string {
	return string(*e)
}

const (
	PodResource  EventResourceType = "Pod"
	HAPResource  EventResourceType = "HPA"
	NodeResource EventResourceType = "Node"
)

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
}
