package incident

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/porter-dev/porter-agent/api/server/config"
	"github.com/porter-dev/porter/api/server/shared"
	"github.com/porter-dev/porter/api/server/shared/apierrors"
	"github.com/porter-dev/porter/api/server/shared/requestutils"
	"gorm.io/gorm"
)

type GetIncidentHandler struct {
	resultWriter shared.ResultWriter
	config       *config.Config
}

func NewGetIncidentHandler(config *config.Config) *GetIncidentHandler {
	return &GetIncidentHandler{
		resultWriter: shared.NewDefaultResultWriter(config.Logger, config.Alerter),
		config:       config,
	}
}

func (h *GetIncidentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	incidentUID, reqErr := requestutils.GetURLParamString(r, "incident_id")

	if reqErr != nil {
		apierrors.HandleAPIError(h.config.Logger, h.config.Alerter, w, r, reqErr, true)
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

	h.resultWriter.WriteResult(w, r, incident.ToAPIType())
}
