package incident

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/porter-dev/porter-agent/api/server/config"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/internal/utils"
	"github.com/porter-dev/porter/api/server/shared"
	"github.com/porter-dev/porter/api/server/shared/apierrors"
	"github.com/porter-dev/porter/api/server/shared/requestutils"
	"gorm.io/gorm"
)

type ListIncidentEventsHandler struct {
	decoderValidator shared.RequestDecoderValidator
	resultWriter     shared.ResultWriter
	config           *config.Config
}

func NewListIncidentEventsHandler(config *config.Config) *ListIncidentEventsHandler {
	return &ListIncidentEventsHandler{
		resultWriter:     shared.NewDefaultResultWriter(config.Logger, config.Alerter),
		decoderValidator: shared.NewDefaultRequestDecoderValidator(config.Logger, config.Alerter),
		config:           config,
	}
}

func (h *ListIncidentEventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	incidentUID, reqErr := requestutils.GetURLParamString(r, "incident_id")

	if reqErr != nil {
		apierrors.HandleAPIError(h.config.Logger, h.config.Alerter, w, r, reqErr, true)
		return
	}

	req := &types.ListIncidentEventsRequest{
		PaginationRequest: &types.PaginationRequest{},
	}

	if ok := h.decoderValidator.DecodeAndValidate(w, r, req); !ok {
		return
	}

	incident, err := h.config.Repository.Incident.ReadIncident(incidentUID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierrors.HandleAPIError(h.config.Logger, h.config.Alerter, w, r,
				apierrors.NewErrNotFound(fmt.Errorf("no such incident exists")), true)
			return
		}

		apierrors.HandleAPIError(h.config.Logger, h.config.Alerter, w, r, apierrors.NewErrInternal(err), true)
		return
	}

	events, paginatedResult, err := h.config.Repository.IncidentEvent.ListEvents(
		&utils.ListIncidentEventsFilter{
			IncidentID:   &incident.ID,
			PodName:      req.PodName,
			PodNamespace: req.PodNamespace,
			Summary:      req.Summary,
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

	res := &types.ListIncidentEventsResponse{
		Pagination: &types.PaginationResponse{
			NumPages:    paginatedResult.NumPages,
			CurrentPage: paginatedResult.CurrentPage,
			NextPage:    paginatedResult.NextPage,
		},
	}

	for _, ev := range events {
		res.Events = append(res.Events, ev.ToAPIType())
	}

	h.resultWriter.WriteResult(w, r, res)
}
