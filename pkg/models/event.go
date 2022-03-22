package models

type EventResourceType string

func (e *EventResourceType) String() string {
	return string(*e)
}

const (
	PodResource EventResourceType = "Pod"
)

type EventCriticality string

type ContainerEvent struct {
	Name     string `json:"container_name"`
	Reason   string `json:"reason"`
	Message  string `json:"message"`
	LogID    string `json:"log_id"`
	ExitCode int32  `json:"exit_code"`
}

type PodEvent struct {
	EventID         string                     `json:"event_id"`
	ChartName       string                     `json:"release_chart_name"`
	PodName         string                     `json:"pod_name"`
	Namespace       string                     `json:"namespace"`
	Cluster         string                     `json:"cluster"`
	OwnerName       string                     `json:"release_name"`
	OwnerType       string                     `json:"release_type"`
	Timestamp       int64                      `json:"timestamp"`
	Phase           string                     `json:"pod_phase"`
	Status          string                     `json:"pod_status"`
	Reason          string                     `json:"reason"`
	Message         string                     `json:"message"`
	ContainerEvents map[string]*ContainerEvent `json:"container_events"`
}
