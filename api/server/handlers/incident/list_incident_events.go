package incident

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/schema"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/internal/repository"
	"github.com/porter-dev/porter-agent/internal/utils"
)

type ListIncidentEventsHandler struct {
	repo *repository.Repository
}

func (h ListIncidentEventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	incidentID := chi.URLParam(r, "uid")

	if incidentID == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("API error in ListIncidentEventsHandler: %v", fmt.Errorf("empty incident id"))
		return
	}

	req := &types.ListIncidentEventsRequest{}

	err := schema.NewDecoder().Decode(req, r.URL.Query())

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("API error in ListIncidentEventsHandler: %v", err)
		return
	}

	events, paginatedResult, err := h.repo.IncidentEvent.ListEvents(
		&utils.ListIncidentEventsFilter{
			IncidentID:   &incidentID,
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
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("API error in ListIncidentEventsHandler: %v", err)
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

	jsonResponse, err := json.Marshal(res)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("API error in ListIncidentEventsHandler: %v", err)
		return
	}

	w.Write(jsonResponse)
}