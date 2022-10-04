package incident

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/schema"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/internal/repository"
	"github.com/porter-dev/porter-agent/internal/utils"
	"gorm.io/gorm"
)

type ListIncidentEventsHandler struct {
	repo *repository.Repository
}

func NewListIncidentEventsHandler(repo *repository.Repository) *ListIncidentEventsHandler {
	return &ListIncidentEventsHandler{repo}
}

func (h ListIncidentEventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	incidentUID := chi.URLParam(r, "uid")

	if incidentUID == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("API error in ListIncidentEventsHandler: %v", fmt.Errorf("empty incident id"))
		return
	}

	req := &types.ListIncidentEventsRequest{
		PaginationRequest: &types.PaginationRequest{},
	}

	err := schema.NewDecoder().Decode(req, r.URL.Query())

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("API error in ListIncidentEventsHandler: %v", err)
		return
	}

	incident, err := h.repo.Incident.ReadIncident(incidentUID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			log.Printf("API error in ListIncidentEventsHandler: %v", err)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		log.Printf("API error in ListIncidentEventsHandler: %v", err)
		return
	}

	events, paginatedResult, err := h.repo.IncidentEvent.ListEvents(
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
