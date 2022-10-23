package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/repository"
)

// ArgoCDResourceHookConsumer contains all information required to consumer an ArgoCDResourceHook
type ArgoCDResourceHookConsumer struct {
	Repository *repository.Repository
}

// NewArgoCDResourceHookConsumer creates an ArgoCDResourceHookConsumer
func NewArgoCDResourceHookConsumer(repo *repository.Repository) ArgoCDResourceHookConsumer {
	return ArgoCDResourceHookConsumer{
		Repository: repo,
	}
}

// Consume contains all business logic for consuming an argo resource hook
func (co ArgoCDResourceHookConsumer) Consume(ctx context.Context, argoEvent types.ArgoCDResourceHook) error {
	co.logger.Caller().Info().Msgf("Received argo hook: %#v\n", argoEvent)

	ty, ok := resourceHookToEvent[argoEvent.Status]
	if !ok {
		return fmt.Errorf("unsupported type %s", argoEvent.Status)
	}

	ti := time.Now().UTC()

	argoBytes, err := json.Marshal(argoEvent)
	if err != nil {
		return fmt.Errorf("error marshalling argo event: %w", err)
	}

	porterEvent := models.Event{
		Type:             ty,
		ReleaseName:      argoEvent.Application,
		ReleaseNamespace: argoEvent.ApplicationNamespace,
		Version:          "v1.0.0",
		Timestamp:        &ti,
		Data:             argoBytes,
		UniqueID:         uuid.NewString(),
	}

	ev, err := co.Repository.Event.CreateEvent(&porterEvent)
	if err != nil {
		return fmt.Errorf("error storing argo event: %w", err)
	}

	fmt.Printf("porterEvent %#v\n\n", ev.UniqueID)

	return nil
}

var resourceHookToEvent = map[string]types.EventType{
	"Synced":    types.EventTypeDeploymentFinished,
	"OutOfSync": types.EventTypeDeploymentStarted,
}
