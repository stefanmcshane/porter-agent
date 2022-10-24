package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/internal/logger"
	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/repository"
)

// ArgoCDResourceHookConsumer contains all information required to consumer an ArgoCDResourceHook
type ArgoCDResourceHookConsumer struct {
	Repository *repository.Repository
	logger     *logger.Logger
}

// NewArgoCDResourceHookConsumer creates an ArgoCDResourceHookConsumer
func NewArgoCDResourceHookConsumer(repo *repository.Repository, logger *logger.Logger) ArgoCDResourceHookConsumer {
	return ArgoCDResourceHookConsumer{
		Repository: repo,
		logger:     logger,
	}
}

// Consume contains all business logic for consuming an argo resource hook
func (co ArgoCDResourceHookConsumer) Consume(ctx context.Context, argoEvent types.ArgoCDResourceHook) error {
	co.logger.Debug().Msgf("Received argo hook: %#v\n", argoEvent)

	ty, ok := resourceHookToEvent[argoEvent.Status]
	if !ok {
		return fmt.Errorf("unsupported type %s", argoEvent.Status)
	}

	ti := time.Now().UTC()

	argoBytes, err := json.Marshal(argoEvent)
	if err != nil {
		return fmt.Errorf("error marshalling argo event: %w", err)
	}

	uid, err := models.GenerateRandomBytes(16)
	if err != nil {
		return fmt.Errorf("error generating random bytes: %w", err)
	}

	porterEvent := models.Event{
		Type:             ty,
		ReleaseName:      argoEvent.Application,
		ReleaseNamespace: argoEvent.ApplicationNamespace,
		Version:          "v1",
		Timestamp:        &ti,
		Data:             argoBytes,
		UniqueID:         uid,
	}

	_, err = co.Repository.Event.CreateEvent(&porterEvent)
	if err != nil {
		return fmt.Errorf("error storing argo event: %w", err)
	}

	return nil
}

var resourceHookToEvent = map[string]types.EventType{
	"Synced":    types.EventTypeDeploymentFinished,
	"OutOfSync": types.EventTypeDeploymentStarted,
}
