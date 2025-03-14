package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	me "github.com/zetcan333/metrics-collector/pkg/myerrors"
)

//go:generate go run github.com/vektra/mockery/v2@v2.52.3 --name=Updater
type ServerUseCase interface {
	UpdateMetric(metricType, metricName, metricValue string) error
	GetValue(metricType, metricName string) (string, error)
	GetAllMetric() (string, error)
}

type ServerHandler struct {
	serverUseCase ServerUseCase
}

func NewServerHandler(suc ServerUseCase) *ServerHandler {
	return &ServerHandler{serverUseCase: suc}
}
func (h *ServerHandler) NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", h.GetAllMetricsHandler)
		r.Route("/update", func(r chi.Router) {
			r.Post("/{type}/{name}/{value}", h.UpdateHandle)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{type}/{name}", h.GetValueHandler)
		})
	})
	return r
}

// UpdateHandler обрабатывает запросы на обновление метрик
func (h *ServerHandler) UpdateHandle(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	if err := h.serverUseCase.UpdateMetric(metricType, metricName, metricValue); err != nil {
		switch {
		case errors.Is(err, me.ErrInvalidMetricType):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, me.ErrInvalidGaugeValue):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, me.ErrInvalidCounterValue):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetValueHandler возвращает значние метрики
func (h *ServerHandler) GetValueHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	res, err := h.serverUseCase.GetValue(metricType, metricName)
	if err != nil {
		switch {
		case errors.Is(err, me.ErrMetricNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, me.ErrInvalidMetricType):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}
	fmt.Fprintf(w, "%s", res)
}

// GetAllMetricsHandler возвращает все метрики в формате HTML
func (h *ServerHandler) GetAllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	html, err := h.serverUseCase.GetAllMetric()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}
