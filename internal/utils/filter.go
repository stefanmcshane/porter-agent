package utils

import "github.com/porter-dev/porter-agent/internal/models"

type ListIncidentsFilter struct {
	Status           *models.IncidentStatus
	ReleaseName      *string
	ReleaseNamespace *string
}

type ListIncidentEventsFilter struct {
	IncidentID     *uint
	PodName        *string
	PodNamespace   *string
	Summary        *string
	IsPrimaryCause *bool
}
