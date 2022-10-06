package incident

import (
	"net/http"

	"github.com/porter-dev/porter-agent/api/server/config"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/internal/utils"
	"github.com/porter-dev/porter/api/server/shared"
	"github.com/porter-dev/porter/api/server/shared/apierrors"
)

type ListIncidentsHandler struct {
	decoderValidator shared.RequestDecoderValidator
	resultWriter     shared.ResultWriter
	config           *config.Config
}

func NewListIncidentsHandler(config *config.Config) *ListIncidentsHandler {
	return &ListIncidentsHandler{
		resultWriter:     shared.NewDefaultResultWriter(config.Logger, config.Alerter),
		decoderValidator: shared.NewDefaultRequestDecoderValidator(config.Logger, config.Alerter),
		config:           config,
	}
}

func (h ListIncidentsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &types.ListIncidentsRequest{
		PaginationRequest: &types.PaginationRequest{},
	}

	if ok := h.decoderValidator.DecodeAndValidate(w, r, req); !ok {
		return
	}

	incidents, paginatedResult, err := h.config.Repository.Incident.ListIncidents(
		&utils.ListIncidentsFilter{
			Status:           req.Status,
			ReleaseName:      req.ReleaseName,
			ReleaseNamespace: req.ReleaseNamespace,
		},
		utils.WithSortBy("updated_at"),
		utils.WithOrder(utils.OrderDesc),
		utils.WithLimit(50),
		utils.WithOffset(req.Page*50),
	)

	if err != nil {
		apierrors.HandleAPIError(h.config.Logger, h.config.Alerter, w, r, apierrors.NewErrInternal(err), true)
		return
	}

	res := &types.ListIncidentsResponse{
		Pagination: &types.PaginationResponse{
			NumPages:    paginatedResult.NumPages,
			CurrentPage: paginatedResult.CurrentPage,
			NextPage:    paginatedResult.NextPage,
		},
	}

	for _, incident := range incidents {
		res.Incidents = append(res.Incidents, incident.ToAPITypeMeta())
	}

	h.resultWriter.WriteResult(w, r, res)
}
