package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type GaugeUpdater interface {
	UpdateGauge(name string, value float64)
}

type GaugeGetter interface {
	GetGauge(name string) (float64, bool)
}

type CounterUpdater interface {
	UpdateCounter(name string, value int64)
}

type CounterGetter interface {
	GetCounter(name string) (int64, bool)
}

// UpdateGaugeHandler обрабатывает запросы на обновление метрик типа gauge.
func UpdateGaugeHandler(gaugeUpdater GaugeUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Разбираем путь запроса
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 5 {
			http.Error(w, "Invalid request path", http.StatusNotFound)
			return
		}

		metricName := parts[3]
		metricValue := parts[4]

		// Проверяем, что имя метрики не пустое
		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		// Парсим значение метрики
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}

		// Обновляем метрику
		gaugeUpdater.UpdateGauge(metricName, value)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Gauge metric updated")
	}
}

// UpdateCounterHandler обрабатывает запросы на обновление метрик типа counter.
func UpdateCounterHandler(counterUpdater CounterUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Разбираем путь запроса
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 5 {
			http.Error(w, "Invalid request path", http.StatusNotFound)
			return
		}

		metricName := parts[3]
		metricValue := parts[4]

		// Проверяем, что имя метрики не пустое
		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		// Парсим значение метрики
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}

		// Обновляем метрику
		counterUpdater.UpdateCounter(metricName, value)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Counter metric updated")
	}
}
