package types

type ListIncidentsRequest struct {
	Status           string `schema:"status"`
	ReleaseName      string `schema:"release_name"`
	ReleaseNamespace string `schema:"release_namespace"`
}

type Incident struct {
	ID               string `json:"id" form:"required"`
	ReleaseName      string `json:"release_name" form:"required"`
	ReleaseNamespace string `json:"release_namespace" form:"required"`
	ChartName        string `json:"chart_name"`
	CreatedAt        int64  `json:"created_at" form:"required"`
	UpdatedAt        int64  `json:"updated_at" form:"required"`
	LatestState      string `json:"latest_state" form:"required"`
	LatestReason     string `json:"latest_reason" form:"required"`
	LatestMessage    string `json:"latest_message" form:"required"`
}
