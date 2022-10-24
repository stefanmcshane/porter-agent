package types

import "time"

type GetKubernetesEventRequest struct {
	Limit       uint       `schema:"limit"`
	StartRange  *time.Time `schema:"start_range"`
	EndRange    *time.Time `schema:"end_range"`
	Revision    string     `schema:"revision"`
	PodSelector string     `schema:"pod_selector" form:"required"`
	Namespace   string     `schema:"namespace" form:"required"`
}

type KubernetesEventLine struct {
	Timestamp *time.Time `json:"timestamp"`
	Event     string     `json:"event"`
}

type GetKubernetesEventResponse struct {
	ContinueTime *time.Time            `json:"continue_time"`
	Events       []KubernetesEventLine `json:"events"`
}
