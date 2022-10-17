package log

import (
	"net/http"
	"time"

	"github.com/porter-dev/porter-agent/api/server/config"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/pkg/logstore"
	"github.com/porter-dev/porter/api/server/shared"
	"github.com/porter-dev/porter/api/server/shared/apierrors"
)

type GetRevisionValuesHandler struct {
	decoderValidator shared.RequestDecoderValidator
	resultWriter     shared.ResultWriter

	Config *config.Config
}

func NewGetRevisionValuesHandler(config *config.Config) *GetRevisionValuesHandler {
	return &GetRevisionValuesHandler{
		resultWriter:     shared.NewDefaultResultWriter(config.Logger, config.Alerter),
		decoderValidator: shared.NewDefaultRequestDecoderValidator(config.Logger, config.Alerter),
		Config:           config,
	}
}

func (h *GetRevisionValuesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &types.GetRevisionValuesRequest{}

	if ok := h.decoderValidator.DecodeAndValidate(w, r, req); !ok {
		return
	}

	if req.StartRange == nil {
		days29 := time.Now().Add(-29 * 24 * time.Hour)
		req.StartRange = &days29
	}

	if req.EndRange == nil {
		now := time.Now()
		req.EndRange = &now
	}

	vals, err := h.Config.LogStore.GetRevisionLabelValues(logstore.LabelValueOptions{
		Start:     *req.StartRange,
		End:       *req.EndRange,
		PodPrefix: req.MatchPrefix,
	})

	if err != nil {
		apierrors.HandleAPIError(h.Config.Logger, h.Config.Alerter, w, r, apierrors.NewErrInternal(err), true)
		return
	}

	h.resultWriter.WriteResult(w, r, vals)
}
