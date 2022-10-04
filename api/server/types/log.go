package types

import "time"

type GetLogRequest struct {
	Limit      uint       `json:"limit"`
	StartRange *time.Time `json:"start_range"`
	EndRange   *time.Time `json:"end_range"`
}

type LogLine struct {
	Timestamp *time.Time `json:"timestamp"`
	Line      string     `json:"line"`
}

type GetLogResponse struct {
	ContinueTime *time.Time `json:"continue_time"`
	Logs         []LogLine  `json:"logs"`
}
