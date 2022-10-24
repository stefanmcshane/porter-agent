package incident

import (
	"fmt"
	"net/http"

	"github.com/porter-dev/porter-agent/api/server/config"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/internal/utils"
	"github.com/porter-dev/porter/api/server/shared"
	"github.com/porter-dev/porter/api/server/shared/apierrors"
)

type ListJobEventsHandler struct {
	decoderValidator shared.RequestDecoderValidator
	resultWriter     shared.ResultWriter
	config           *config.Config
}

func NewListJobEventsHandler(config *config.Config) *ListJobEventsHandler {
	return &ListJobEventsHandler{
		resultWriter:     shared.NewDefaultResultWriter(config.Logger, config.Alerter),
		decoderValidator: shared.NewDefaultRequestDecoderValidator(config.Logger, config.Alerter),
		config:           config,
	}
}

func (h ListJobEventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &types.ListJobEventsRequest{
		PaginationRequest: &types.PaginationRequest{},
	}

	if ok := h.decoderValidator.DecodeAndValidate(w, r, req); !ok {
		return
	}

	queryMeta := fmt.Sprintf("job/%s", req.JobName)

	events, paginatedResult, err := h.config.Repository.Event.ListEvents(
		&utils.ListEventsFilter{
			ReleaseName:         req.ReleaseName,
			ReleaseNamespace:    req.ReleaseNamespace,
			Type:                req.Type,
			AdditionalQueryMeta: &queryMeta,
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
