package types

import "github.com/porter-dev/porter-agent/internal/models"

type ListIncidentsRequest struct {
	Status           models.IncidentStatus `schema:"status"`
	ReleaseName      string                `schema:"release_name"`
	ReleaseNamespace string                `schema:"release_namespace"`
}
