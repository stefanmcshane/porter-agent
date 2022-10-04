package types

import "time"

type GetLogRequest struct {
	Limit       uint       `schema:"limit"`
	StartRange  *time.Time `schema:"start_range"`
	EndRange    *time.Time `schema:"end_range"`
	PodSelector string     `schema:"pod_selector" form:"required"`
}

type LogLine struct {
	Timestamp *time.Time `json:"timestamp"`
	Line      string     `json:"line"`
}

type GetLogResponse struct {
	ContinueTime *time.Time `json:"continue_time"`
	Logs         []LogLine  `json:"logs"`
}
