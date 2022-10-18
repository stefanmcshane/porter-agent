package incident

import (
	"net/http"

	"github.com/porter-dev/porter-agent/api/server/config"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/internal/utils"
	"github.com/porter-dev/porter/api/server/shared"
	"github.com/porter-dev/porter/api/server/shared/apierrors"
)

type ListEventsHandler struct {
	decoderValidator shared.RequestDecoderValidator
	resultWriter     shared.ResultWriter
	config           *config.Config
}

func NewListEventsHandler(config *config.Config) *ListEventsHandler {
	return &ListEventsHandler{
		resultWriter:     shared.NewDefaultResultWriter(config.Logger, config.Alerter),
		decoderValidator: shared.NewDefaultRequestDecoderValidator(config.Logger, config.Alerter),
		config:           config,
	}
}

func (h ListEventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &types.ListEventsRequest{
		PaginationRequest: &types.PaginationRequest{},
	}

	if ok := h.decoderValidator.DecodeAndValidate(w, r, req); !ok {
		return
	}

	events, paginatedResult, err := h.config.Repository.Event.ListEvents(
		&utils.ListEventsFilter{
			ReleaseName:      req.ReleaseName,
			ReleaseNamespace: req.ReleaseNamespace,
			Type:             req.Type,
		},
		utils.WithSortBy("timestamp"),
		utils.WithOrder(utils.OrderDesc),
		utils.WithLimit(50),
		utils.WithOffset(req.Page*50),
	)

	if err != nil {
		apierrors.HandleAPIError(h.config.Logger, h.config.Alerter, w, r, apierrors.NewErrInternal(err), true)
		return
	}

	res := &types.ListEventsResponse{
		Pagination: &types.PaginationResponse{
			NumPages:    paginatedResult.NumPages,
			CurrentPage: paginatedResult.CurrentPage,
			NextPage:    paginatedResult.NextPage,
		},
	}

	for _, event := range events {
		res.Events = append(res.Events, event.ToAPIType())
	}

	h.resultWriter.WriteResult(w, r, res)
}
