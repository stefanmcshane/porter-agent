package utils

import "github.com/porter-dev/porter-agent/api/server/types"

type ListIncidentsFilter struct {
	Status           *types.IncidentStatus
	ReleaseName      *string
	ReleaseNamespace *string
}

type ListIncidentEventsFilter struct {
	IncidentID     *string
	PodName        *string
	PodNamespace   *string
	Summary        *string
	IsPrimaryCause *bool
}
