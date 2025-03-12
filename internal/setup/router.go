package setup

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handlers interface {
	UpdateHandle(w http.ResponseWriter, r *http.Request)
	GetValueHandler(w http.ResponseWriter, r *http.Request)
	GetAllMetricsHandler(w http.ResponseWriter, r *http.Request)
}

type Routes struct {
	handlers Handlers
}

func NewSetup(h Handlers) *Routes {
	return &Routes{handlers: h}
}

func (rr *Routes) SetRoutes(r *chi.Mux) {
	r.Route("/", func(r chi.Router) {
		r.Get("/", rr.handlers.GetAllMetricsHandler)
		r.Route("/update", func(r chi.Router) {
			r.Post("/{type}/{name}/{value}", rr.handlers.UpdateHandle)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{type}/{name}", rr.handlers.GetValueHandler)
		})
	})
}
