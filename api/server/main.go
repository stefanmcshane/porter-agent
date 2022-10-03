package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/porter-dev/porter-agent/api/server/handlers/incident"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Method("GET", "/incidents", &incident.ListIncidentsHandler{})
	r.Method("GET", "/incidents/{uid}", &incident.GetIncidentHandler{})
	r.Method("GET", "/incidents/{uid}/events", &incident.ListIncidentEventsHandler{})

	http.ListenAndServe(":3000", r)
}
