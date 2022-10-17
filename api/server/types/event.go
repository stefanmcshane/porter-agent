package types

import "time"

type GetEventRequest struct {
	Limit       uint       `schema:"limit"`
	StartRange  *time.Time `schema:"start_range"`
	EndRange    *time.Time `schema:"end_range"`
	Revision    string     `schema:"revision"`
	PodSelector string     `schema:"pod_selector" form:"required"`
	Namespace   string     `schema:"namespace" form:"required"`
}

type EventLine struct {
	Timestamp *time.Time `json:"timestamp"`
	Event     string     `json:"event"`
}

type GetEventResponse struct {
	ContinueTime *time.Time  `json:"continue_time"`
	Events       []EventLine `json:"events"`
}
