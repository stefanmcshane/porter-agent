package incident

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/internal/repository"
	"github.com/porter-dev/porter-agent/internal/utils"
)

type ListIncidentsHandler struct {
	repo *repository.Repository
}

func (h ListIncidentsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &types.ListIncidentsRequest{}

	err := schema.NewDecoder().Decode(req, r.URL.Query())

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("API error in ListIncidentsHandler: %v", err)
		return
	}

	incidents, paginatedResult, err := h.repo.Incident.ListIncidents(
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
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("API error in ListIncidentsHandler: %v", err)
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

	jsonResponse, err := json.Marshal(res)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("API error in ListIncidentsHandler: %v", err)
		return
	}

	w.Write(jsonResponse)
}
