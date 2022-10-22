package healthcheck

import (
	"fmt"
	"net/http"

	"github.com/porter-dev/porter-agent/api/server/config"
	"github.com/porter-dev/porter/api/server/shared"
	"github.com/porter-dev/porter/api/server/shared/apierrors"
)

type ReadyzHandler struct {
	resultWriter shared.ResultWriter
	config       *config.Config
}

func NewReadyzHandler(config *config.Config) *ReadyzHandler {
	return &ReadyzHandler{
		resultWriter: shared.NewDefaultResultWriter(config.Logger, config.Alerter),
		config:       config,
	}
}

func (h *ReadyzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db := h.config.Repository.DB

	switch db.Dialector.Name() {
	case "sqlite":
		writeHealthy(w)
		return
	case "postgres":
		sqlDB, err := db.DB()

		if err != nil {
			apierrors.HandleAPIError(h.config.Logger, h.config.Alerter, w, r, apierrors.NewErrInternal(err), true)
			return
		}

		if err := sqlDB.Ping(); err != nil {
			apierrors.HandleAPIError(h.config.Logger, h.config.Alerter, w, r, apierrors.NewErrInternal(err), true)
			return
		}

		writeHealthy(w)
		return
	}

	apierrors.HandleAPIError(h.config.Logger, h.config.Alerter, w, r, apierrors.NewErrPassThroughToClient(
		fmt.Errorf("database is not supported"),
		http.StatusBadRequest,
	), true)

	return
}

func writeHealthy(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("."))
}
