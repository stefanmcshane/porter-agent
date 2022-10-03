package incident

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/porter-dev/porter-agent/internal/repository"
	"gorm.io/gorm"
)

type GetIncidentHandler struct {
	repo *repository.Repository
}

func NewGetIncidentHandler(repo *repository.Repository) *GetIncidentHandler {
	return &GetIncidentHandler{repo}
}

func (h GetIncidentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	incidentUID := chi.URLParam(r, "uid")

	if incidentUID == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("API error in GetIncidentHandler: %v", fmt.Errorf("empty incident id"))
		return
	}

	incident, err := h.repo.Incident.ReadIncident(incidentUID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			log.Printf("API error in GetIncidentHandler: %v", err)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		log.Printf("API error in GetIncidentHandler: %v", err)
		return
	}

	jsonResponse, err := json.Marshal(incident.ToAPIType())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("API error in GetIncidentHandler: %v", err)
		return
	}

	w.Write(jsonResponse)
}
