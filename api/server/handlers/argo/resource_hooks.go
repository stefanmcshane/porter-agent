package argo

import (
	"fmt"
	"net/http"

	"github.com/porter-dev/porter-agent/api/server/config"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/pkg/controllers"
	"github.com/porter-dev/porter/api/server/shared"
)

// ResourceHookHandler contains helper functions for listening to Argo CD Resource Hook events
type ResourceHookHandler struct {
	decoderValidator shared.RequestDecoderValidator
	resultWriter     shared.ResultWriter

	Config                 *config.Config
	ArgoCDResourceConsumer controllers.ArgoCDResourceHookConsumer
}

// NewResourceHookHandler returns a new instance of ResourceHookHandler
func NewResourceHookHandler(config *config.Config, consumer controllers.ArgoCDResourceHookConsumer) ResourceHookHandler {
	return ResourceHookHandler{
		resultWriter:           shared.NewDefaultResultWriter(config.Logger, config.Alerter),
		decoderValidator:       shared.NewDefaultRequestDecoderValidator(config.Logger, config.Alerter),
		Config:                 config,
		ArgoCDResourceConsumer: consumer,
	}
}

// ServeHTTP implements Go's HTTP handler interface for listening to ArgoCD resource events
func (h ResourceHookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req types.ArgoCDResourceHook

	if ok := h.decoderValidator.DecodeAndValidate(w, r, &req); !ok {
		return
	}

	err := h.ArgoCDResourceConsumer.Consume(r.Context(), req)
	if err != nil {
		fmt.Println("error", err)
		return
	}
}
