package application

import (
	"net/http"

	"github.com/porter-dev/porter-agent/api/server/config"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/pkg/argocd"
	"github.com/porter-dev/porter/api/server/shared"
	"github.com/porter-dev/porter/api/server/shared/apierrors"
)

// ApplicationRollbackHandler contains helper functions for listening to application rollback requests
type ApplicationRollbackHandler struct {
	decoderValidator shared.RequestDecoderValidator
	resultWriter     shared.ResultWriter
	config           *config.Config

	ArgoCD argocd.ArgoCD
}

// NewApplicationRollbackHandler returns a new instance of ApplicationRollbackHandler
func NewApplicationRollbackHandler(config *config.Config, consumer argocd.ArgoCD) ApplicationRollbackHandler {
	return ApplicationRollbackHandler{
		resultWriter:     shared.NewDefaultResultWriter(config.Logger, config.Alerter),
		decoderValidator: shared.NewDefaultRequestDecoderValidator(config.Logger, config.Alerter),
		config:           config,
		ArgoCD:           consumer,
	}
}

// ServeHTTP implements Go's HTTP handler interface for listening to application sync requests
func (h ApplicationRollbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req types.Application

	if ok := h.decoderValidator.DecodeAndValidate(w, r, &req); !ok {
		return
	}

	application := argocd.Application{
		Name:      req.Name,
		Namespace: req.Namespace,
		Revision:  req.Revision,
		Status:    req.Status,
	}

	err := h.ArgoCD.Rollback(r.Context(), application)
	if err != nil {
		apierrors.HandleAPIError(h.config.Logger, h.config.Alerter, w, r, apierrors.NewErrInternal(err), true)
		return
	}
}
