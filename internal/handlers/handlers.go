package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
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
type AllGetter interface {
	GetAllCounter() map[string]int64
	GetAllGauge() map[string]float64
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

// GetValueHandler возвращает значние метрики
func GetValueHandler(getter Getter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 4 {
			http.Error(w, "Invalid request path", http.StatusNotFound)
			return
		}
		metricType := parts[2]
		metricName := parts[3]

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

// GetAllMetricsHandler возвращает все метрики в формате HTML
func GetAllMetricsHandler(getter AllGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := `<html>
		<head><title>Metrics</title></head>
		<body>
			<h1>Metrics</h1>
			<ul>
			{{ range $key, $value := .Gauges }}
				<li>Gauge {{$key}}: {{ printf "%.2f" $value }}</li>
			{{ end }}
			{{ range $key, $value := .Counters }}
				<li>Counter {{$key}}: {{$value}}</li>
			{{ end }}
			</ul>
		</body>
		</html>`
		data := struct {
			Gauges   map[string]float64
			Counters map[string]int64
		}{
			Gauges:   getter.GetAllGauge(),
			Counters: getter.GetAllCounter(),
		}
		t, err := template.New("metrics").Parse(tmpl)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		t.Execute(w, data)
	}
}
