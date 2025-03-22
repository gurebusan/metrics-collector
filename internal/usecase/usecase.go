package usecase

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/zetcan333/metrics-collector/internal/lib/format/float"
	"github.com/zetcan333/metrics-collector/internal/models"
	me "github.com/zetcan333/metrics-collector/pkg/myerrors"
)

type ServerRepository interface {
	UpdateMetric(metric models.Metrics)
	GetMetric(id string) (models.Metrics, bool)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

type SeverUsecase struct {
	repo ServerRepository
}

func NewSeverUsecase(repo ServerRepository) *SeverUsecase {
	return &SeverUsecase{repo: repo}
}

func (s *SeverUsecase) UpdateMetric(metricType, metricName, metricValue string) error {
	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("%w: %v", me.ErrInvalidGaugeValue, err)
		}
		s.repo.UpdateMetric(models.Metrics{
			MType: "gauge",
			ID:    metricName,
			Value: &value,
		})

	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %v", me.ErrInvalidCounterValue, err)
		}
		s.repo.UpdateMetric(models.Metrics{
			MType: "counter",
			ID:    metricName,
			Delta: &value,
		})

	default:
		return me.ErrInvalidMetricType
	}
	return nil
}

func (s *SeverUsecase) GetValue(metricType, metricName string) (string, error) {
	metric, ok := s.repo.GetMetric(metricName)
	if !ok {
		return "", me.ErrMetricNotFound
	}

	switch metricType {
	case "gauge":
		if metric.Value == nil {
			return "", me.ErrMetricNotFound
		}
		return float.FormatFloat(*metric.Value), nil

	case "counter":
		if metric.Delta == nil {
			return "", me.ErrMetricNotFound
		}
		return fmt.Sprintf("%d", *metric.Delta), nil

	default:
		return "", me.ErrInvalidMetricType
	}
}

func (s *SeverUsecase) UpdateJSON(metric models.Metrics) (models.Metrics, error) {
	if metric.MType != "gauge" && metric.MType != "counter" {
		return models.Metrics{}, me.ErrInvalidMetricType
	}
	switch metric.MType {
	case "gauge":
		if metric.Value == nil {
			return models.Metrics{}, me.ErrInvalidGaugeValue
		}
		s.repo.UpdateMetric(metric)

		updatedMetric, exists := s.repo.GetMetric(metric.ID)
		if !exists {
			return models.Metrics{}, me.ErrMetricNotFound
		}
		return updatedMetric, nil
	case "counter":
		if metric.Delta == nil {
			return models.Metrics{}, me.ErrInvalidCounterValue
		}
		s.repo.UpdateMetric(metric)

		updatedMetric, exists := s.repo.GetMetric(metric.ID)
		if !exists {
			return models.Metrics{}, me.ErrMetricNotFound
		}
		return updatedMetric, nil
	}
	return models.Metrics{}, me.ErrInvalidMetricType
}

func (s *SeverUsecase) GetJSON(metric models.Metrics) (models.Metrics, error) {

	if metric.MType != "gauge" && metric.MType != "counter" {
		return models.Metrics{}, me.ErrInvalidMetricType
	}
	if metric.ID == "" {
		return models.Metrics{}, me.ErrMetricNotFound
	}

	storedMetric, exists := s.repo.GetMetric(metric.ID)
	if !exists {
		return models.Metrics{}, me.ErrMetricNotFound
	}

	return storedMetric, nil
}

func (s *SeverUsecase) GetAllMetric() (string, error) {
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
		Gauges:   s.repo.GetAllGauges(),
		Counters: s.repo.GetAllCounters(),
	}
	t, err := template.New("metrics").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
