package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

//go:generate go run github.com/vektra/mockery/v2@v2.52.3 --name=Updater
type Updater interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}

type Getter interface {
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
}

// UpdateHandler обрабатывает запросы на обновление метрик
func UpdateHandler(updater Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")
		metricValue := chi.URLParam(r, "value")

		// Проверяем, что имя метрики не пустое
		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}
		// Обрабатываем метрику в зависимости от её типа
		switch metricType {
		case "gauge":
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Invalid gauge value", http.StatusBadRequest)
				return
			}
			updater.UpdateGauge(metricName, value)
		case "counter":
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Invalid counter value", http.StatusBadRequest)
				return
			}
			updater.UpdateCounter(metricName, value)
		default:
			// Возвращаем 400 для некорректного типа метрики
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)

	}
}

// GetValueHandler возвращает значние метрики
func GetValueHandler(getter Getter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")

		switch metricType {
		case "gauge":
			value, ok := getter.GetGauge(metricName)
			if !ok {
				http.Error(w, "Metric not found", http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, "%.2f", value)
		case "counter":
			value, ok := getter.GetCounter(metricName)
			if !ok {
				http.Error(w, "Metric not found", http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, "%d", value)
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
		}
	}
}

//GetAllHandler
