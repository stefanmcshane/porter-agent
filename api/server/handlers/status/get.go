package status

import (
	"net/http"

	"github.com/porter-dev/porter-agent/api/server/config"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/pkg/logstore/lokistore"
	"github.com/porter-dev/porter/api/server/shared"
)

type GetStatusHandler struct {
	resultWriter shared.ResultWriter
}

func NewGetStatusHandler(config *config.Config) *GetStatusHandler {
	return &GetStatusHandler{
		resultWriter: shared.NewDefaultResultWriter(config.Logger, config.Alerter),
	}
}

func (h *GetStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.resultWriter.WriteResult(w, r, &types.GetStatusResponse{
		Loki: lokistore.GetLokiStatus(),
	})
}
