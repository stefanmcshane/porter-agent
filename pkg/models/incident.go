package models

type Incident struct {
	ID            string `json:"id" form:"required"`
	ReleaseName   string `json:"release_name" form:"required"`
	LatestState   string `json:"latest_state" form:"required"`
	LatestReason  string `json:"latest_reason" form:"required"`
	LatestMessage string `json:"latest_message" form:"required"`
}
