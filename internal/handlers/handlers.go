package handlers

import (
	"net/http"
	"strconv"
	"strings"
)

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
		// Разбираем путь запроса
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 5 {
			http.Error(w, "Invalid request path", http.StatusNotFound)
			return
		}
		metricType := parts[2]
		metricName := parts[3]
		metricValue := parts[4]

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
