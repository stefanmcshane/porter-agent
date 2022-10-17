package types

import "time"

type GetLogRequest struct {
	Limit       uint       `schema:"limit"`
	StartRange  *time.Time `schema:"start_range"`
	EndRange    *time.Time `schema:"end_range"`
	SearchParam string     `schema:"search_param"`
	Revision    string     `schema:"revision"`
	PodSelector string     `schema:"pod_selector" form:"required"`
	Namespace   string     `schema:"namespace" form:"required"`
	Direction   string     `schema:"direction"`
}

type LogLine struct {
	Timestamp *time.Time `json:"timestamp"`
	Line      string     `json:"line"`
}

type GetLogResponse struct {
	BackwardContinueTime *time.Time `json:"backward_continue_time"`
	ForwardContinueTime  *time.Time `json:"forward_continue_time"`
	Logs                 []LogLine  `json:"logs"`
}

type GetPodValuesRequest struct {
	StartRange  *time.Time `schema:"start_range"`
	EndRange    *time.Time `schema:"end_range"`
	MatchPrefix string     `schema:"match_prefix"`
}

type GetRevisionValuesRequest struct {
	StartRange *time.Time `schema:"start_range"`
	EndRange   *time.Time `schema:"end_range"`
}
