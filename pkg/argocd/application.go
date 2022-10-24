package argocd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/stefanmcshane/go-argocd/argocd/client/application_service"
	"github.com/stefanmcshane/go-argocd/argocd/models"
)

// Application wraps an ArgoCD application
type Application struct {
	Name      string
	Namespace string
	Status    string

	// Revision is the revision that an application is currently at.
	// If this parameter is set when running `Rollback`, ArgoCD will attempt to sync the specific revision
	Revision string
}

// Sync will attempt to sync a given ArgoCD application manually
// Doing this will not trigger any webhooks to fire for event notifications
// If no namespace is provided, the 'default' namespace will be assumed
func (a argoCDClient) Sync(ctx context.Context, app Application) error {
	body := models.ApplicationApplicationSyncRequest{
		Name:         app.Name,
		AppNamespace: app.Namespace,
		Prune:        true,
	}

	if body.Name == "" {
		return errors.New("must supply application name to sync")
	}
	if body.AppNamespace == "" {
		body.AppNamespace = "default"
	}
	if app.Revision != "" {
		body.Revision = app.Revision
	}

	params := application_service.NewApplicationServiceSyncParamsWithContext(ctx)
	params.SetBody(&body)
	params.Name = app.Name

	_, err := a.client.ApplicationService.ApplicationServiceSync(params)
	if err != nil {
		return fmt.Errorf("failure when syncing application: %w", err)
	}

	return nil
}

// Rollback will attempt to set a given ArgoCD application to the supplied revision
// If no namespace is provided, the 'default' namespace will be assumed
func (a argoCDClient) Rollback(ctx context.Context, app Application) error {
	body := models.ApplicationApplicationRollbackRequest{
		Name:         app.Name,
		AppNamespace: app.Namespace,
		ID:           app.Revision,
		Prune:        true,
	}

	if body.Name == "" {
		return errors.New("must supply application name to sync")
	}
	if body.ID != "" {
		return errors.New("must set a revision to rollback to")
	}
	if body.AppNamespace == "" {
		body.AppNamespace = "default"
	}

	params := application_service.NewApplicationServiceRollbackParamsWithContext(ctx)
	params.SetBody(&body)
	params.Name = app.Name

	_, err := a.client.ApplicationService.ApplicationServiceRollback(params)
	if err != nil {
		if strings.Contains(err.Error(), "rollback cannot be initiated when auto-sync is enabled") {
			return fmt.Errorf("must disable auto-sync before rolling back")
		}
		return fmt.Errorf("failure when rolling back application: %w", err)
	}

	return nil
}
