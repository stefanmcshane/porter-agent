package incident

import (
	"net/http"

	"github.com/porter-dev/porter-agent/internal/repository"
)

type GetIncidentHandler struct {
	repo *repository.Repository
}

func (h GetIncidentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
