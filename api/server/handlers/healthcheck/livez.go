package healthcheck

import (
	"net/http"

	"github.com/porter-dev/porter-agent/api/server/config"
	"github.com/porter-dev/porter/api/server/shared"
)

type LivezHandler struct {
	resultWriter shared.ResultWriter
	config       *config.Config
}

func NewLivezHandler(config *config.Config) *LivezHandler {
	return &LivezHandler{
		resultWriter: shared.NewDefaultResultWriter(config.Logger, config.Alerter),
		config:       config,
	}
}

func (h *LivezHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeHealthy(w)
}
