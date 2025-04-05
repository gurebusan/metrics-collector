package usecase

import (
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/zetcan333/metrics-collector/internal/lib/format/float"
	"github.com/zetcan333/metrics-collector/internal/models"
	"github.com/zetcan333/metrics-collector/pkg/myerrors"
)

type ServerRepository interface {
	UpdateMetric(ctx context.Context, metric models.Metrics) error
	GetMetric(ctx context.Context, id string) (models.Metrics, error)
	GetAllGauges(ctx context.Context) (map[string]float64, error)
	GetAllCounters(ctx context.Context) (map[string]int64, error)
	UpdateMetricsWithBatch(ctx context.Context, metrics []models.Metrics) error
}

type SeverUsecase struct {
	repo ServerRepository
}

func NewSeverUsecase(repo ServerRepository) *SeverUsecase {
	return &SeverUsecase{repo: repo}
}

var ctx = context.Background()

func (s *SeverUsecase) UpdateMetric(metricType, metricName, metricValue string) error {
	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("%w: %v", myerrors.ErrInvalidGaugeValue, err)
		}
		s.repo.UpdateMetric(ctx, models.Metrics{
			MType: "gauge",
			ID:    metricName,
			Value: &value,
		})

	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %v", myerrors.ErrInvalidCounterValue, err)
		}
		s.repo.UpdateMetric(ctx, models.Metrics{
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
	metric, err := s.repo.GetMetric(ctx, metricName)
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

func (s *SeverUsecase) UpdateViaModel(metric models.Metrics) (models.Metrics, error) {
	if metric.MType != "gauge" && metric.MType != "counter" {
		return models.Metrics{}, myerrors.ErrInvalidMetricType
	}
	switch metric.MType {

	case "gauge":
		if metric.Value == nil {
			return models.Metrics{}, myerrors.ErrInvalidGaugeValue
		}
		s.repo.UpdateMetric(ctx, metric)

		updatedMetric, err := s.repo.GetMetric(ctx, metric.ID)
		if err != nil {
			return models.Metrics{}, err
		}
		return updatedMetric, nil

	case "counter":
		if metric.Delta == nil {
			return models.Metrics{}, myerrors.ErrInvalidCounterValue
		}

		s.repo.UpdateMetric(ctx, metric)

		updatedMetric, err := s.repo.GetMetric(ctx, metric.ID)
		if err != nil {
			return models.Metrics{}, err
		}
		return updatedMetric, nil
	}

	return models.Metrics{}, myerrors.ErrInvalidMetricType
}

func (s *SeverUsecase) GetViaModel(metric models.Metrics) (models.Metrics, error) {

	if metric.MType != "gauge" && metric.MType != "counter" {
		return models.Metrics{}, myerrors.ErrInvalidMetricType
	}

	storedMetric, err := s.repo.GetMetric(ctx, metric.ID)
	if err != nil {
		return models.Metrics{}, err
	}

	return storedMetric, nil
}

func (s *SeverUsecase) GetAllMetrics() (string, error) {
	gauges, err := s.repo.GetAllGauges(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get all gauges: %w", err)
	}
	counters, err := s.repo.GetAllCounters(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get all counters: %w", err)
	}
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
		Gauges:   gauges,
		Counters: counters,
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

func (s *SeverUsecase) UpdateMetricsWithBatch(metrics []models.Metrics) error {
	for _, metric := range metrics {
		if metric.MType != "gauge" && metric.MType != "counter" {
			return myerrors.ErrInvalidMetricType
		}

	}
	return s.repo.UpdateMetricsWithBatch(ctx, metrics)
}
