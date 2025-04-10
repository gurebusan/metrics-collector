package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/zetcan333/metrics-collector/internal/models"
	"github.com/zetcan333/metrics-collector/pkg/myerrors"
	"go.uber.org/zap"
)

//go:generate go run github.com/vektra/mockery/v2@v2.52.3 --name=ServerUseCase
type ServerUseCase interface {
	UpdateMetric(metricType, metricName, metricValue string) error
	GetMetric(metricType, metricName string) (string, error)
	UpdateViaModel(metric models.Metrics) (models.Metrics, error)
	GetViaModel(metric models.Metrics) (models.Metrics, error)
	GetAllMetrics() (string, error)
	UpdateMetricsWithBatch(metrics []models.Metrics) error
}

type ServerHandler struct {
	log           *zap.Logger
	serverUseCase ServerUseCase
}

func NewServerHandler(log *zap.Logger, suc ServerUseCase) *ServerHandler {
	return &ServerHandler{
		log:           log,
		serverUseCase: suc,
	}
}

// UpdateMetric обрабатывает запросы на обновление метрик
func (h *ServerHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if metricName == "" {
		http.Error(w, "metric name is required", http.StatusNotFound)
		return
	}

	if err := h.serverUseCase.UpdateMetric(metricType, metricName, metricValue); err != nil {
		switch {
		case isBadRequest(err):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		default:
			h.log.Sugar().Errorln("falied to update metric", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// GetMetric возвращает значние метрики
func (h *ServerHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	res, err := h.serverUseCase.GetMetric(metricType, metricName)
	if err != nil {
		switch {
		case errors.Is(err, myerrors.ErrMetricNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, myerrors.ErrInvalidMetricType):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			h.log.Sugar().Errorln("falied to get metric", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}
	fmt.Fprintf(w, "%s", res)
}

func (h *ServerHandler) UpdateViaModel(w http.ResponseWriter, r *http.Request) {
	var metric models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		jsonResponse(w, r, http.StatusBadRequest, Error("invalid JSON format"))
		return
	}

	updatedMetric, err := h.serverUseCase.UpdateViaModel(metric)
	if err != nil {
		switch {
		case isBadRequest(err):
			jsonResponse(w, r, http.StatusBadRequest, Error(err.Error()))
			return
		default:
			h.log.Sugar().Errorln("falied to update metric", err)
			jsonResponse(w, r, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	jsonResponse(w, r, http.StatusOK, updatedMetric)
}

func (h *ServerHandler) GetViaModel(w http.ResponseWriter, r *http.Request) {
	var metric models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	result, err := h.serverUseCase.GetViaModel(metric)
	if err != nil {
		switch {
		case errors.Is(err, myerrors.ErrMetricNotFound):
			jsonResponse(w, r, http.StatusNotFound, Error(err.Error()))
			return
		case errors.Is(err, myerrors.ErrInvalidMetricType):
			jsonResponse(w, r, http.StatusBadRequest, Error(err.Error()))
			return
		default:
			h.log.Sugar().Errorln("falied to get metric", err)
			jsonResponse(w, r, http.StatusInternalServerError, Error("internal server error"))
			return
		}
	}

	jsonResponse(w, r, http.StatusOK, result)
}

// GetAllMetricsHandler возвращает все метрики в формате HTML
func (h *ServerHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	html, err := h.serverUseCase.GetAllMetrics()
	if err != nil {
		h.log.Sugar().Errorln("falied to get metrics", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (h *ServerHandler) UpdateMetricsWithBatch(w http.ResponseWriter, r *http.Request) {
	var metrics []models.Metrics
	if err := render.DecodeJSON(r.Body, &metrics); err != nil {
		jsonResponse(w, r, http.StatusBadRequest, Error("invalid JSON format"))
		return
	}

	if len(metrics) == 0 {
		jsonResponse(w, r, http.StatusBadRequest, Error("empty batch"))
		return
	}

	if err := h.serverUseCase.UpdateMetricsWithBatch(metrics); err != nil {
		if isBadRequest(err) {
			jsonResponse(w, r, http.StatusBadRequest, Error(err.Error()))
		} else {
			h.log.Error("failed to update metrics", zap.Error(err))
			jsonResponse(w, r, http.StatusInternalServerError, Error("internal server error"))
		}
		return
	}

	jsonResponse(w, r, http.StatusOK, map[string]string{"status": "batch updated successfully"})
}

func isBadRequest(err error) bool {
	return errors.Is(err, myerrors.ErrInvalidMetricType) ||
		errors.Is(err, myerrors.ErrInvalidGaugeValue) ||
		errors.Is(err, myerrors.ErrInvalidCounterValue)
}

func jsonResponse(w http.ResponseWriter, r *http.Request, StatusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(StatusCode)
	render.JSON(w, r, v)
}

type errorResponse struct {
	Error string `json:"error"`
}

func Error(msg string) errorResponse {
	return errorResponse{Error: msg}
}
