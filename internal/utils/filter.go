package utils

import "github.com/porter-dev/porter-agent/api/server/types"

type ListIncidentsFilter struct {
	Status           *types.IncidentStatus
	ReleaseName      *string
	ReleaseNamespace *string
	Revision         *string
}

type ListEventsFilter struct {
	ReleaseName         *string
	ReleaseNamespace    *string
	Type                *string
	AdditionalQueryMeta *string
}
type ListIncidentEventsFilter struct {
	IncidentID     *uint
	PodName        *string
	PodNamespace   *string
	Summary        *string
	PodPrefix      *string
	IsPrimaryCause *bool
}
