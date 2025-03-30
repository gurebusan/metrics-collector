package usecase

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/zetcan333/metrics-collector/internal/lib/format/float"
	"github.com/zetcan333/metrics-collector/internal/models"
	"github.com/zetcan333/metrics-collector/pkg/myerrors"
)

type ServerRepository interface {
	UpdateMetric(metric models.Metrics)
	GetMetric(id string) (models.Metrics, error)
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
			return fmt.Errorf("%w: %v", myerrors.ErrInvalidGaugeValue, err)
		}
		s.repo.UpdateMetric(models.Metrics{
			MType: "gauge",
			ID:    metricName,
			Value: &value,
		})

	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %v", myerrors.ErrInvalidCounterValue, err)
		}
		s.repo.UpdateMetric(models.Metrics{
			MType: "counter",
			ID:    metricName,
			Delta: &value,
		})

	default:
		return myerrors.ErrInvalidMetricType
	}
	return nil
}

func (s *SeverUsecase) GetMetric(metricType, metricName string) (string, error) {
	metric, err := s.repo.GetMetric(metricName)
	if err != nil {
		return "", err
	}

	switch metricType {
	case "gauge":
		return float.FormatFloat(*metric.Value), nil

	case "counter":
		return fmt.Sprintf("%d", *metric.Delta), nil

	default:
		return "", myerrors.ErrInvalidMetricType
	}
}

func (s *SeverUsecase) UpdateMetric2(metric models.Metrics) (models.Metrics, error) {
	if metric.MType != "gauge" && metric.MType != "counter" {
		return models.Metrics{}, myerrors.ErrInvalidMetricType
	}
	switch metric.MType {

	case "gauge":
		if metric.Value == nil {
			return models.Metrics{}, myerrors.ErrInvalidGaugeValue
		}
		s.repo.UpdateMetric(metric)

		updatedMetric, err := s.repo.GetMetric(metric.ID)
		if err != nil {
			return models.Metrics{}, err
		}
		return updatedMetric, nil

	case "counter":
		if metric.Delta == nil {
			return models.Metrics{}, myerrors.ErrInvalidCounterValue
		}

		s.repo.UpdateMetric(metric)

		updatedMetric, err := s.repo.GetMetric(metric.ID)
		if err != nil {
			return models.Metrics{}, err
		}
		return updatedMetric, nil
	}

	return models.Metrics{}, myerrors.ErrInvalidMetricType
}

func (s *SeverUsecase) GetMetric2(metric models.Metrics) (models.Metrics, error) {

	if metric.MType != "gauge" && metric.MType != "counter" {
		return models.Metrics{}, myerrors.ErrInvalidMetricType
	}

	storedMetric, err := s.repo.GetMetric(metric.ID)
	if err != nil {
		return models.Metrics{}, err
	}

	return storedMetric, nil
}

func (s *SeverUsecase) GetAllMetrics() (string, error) {
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
