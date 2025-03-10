package usercase

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/zetcan333/metrics-collector/internal/lib/format/float"
	me "github.com/zetcan333/metrics-collector/pkg/myerrors"
)

type ServerRepository interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllCounter() map[string]int64
	GetAllGauge() map[string]float64
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
		s.repo.UpdateGauge(metricName, value)
	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %v", me.ErrInvalidCounterValue, err)
		}
		s.repo.UpdateCounter(metricName, value)
	default:

		return me.ErrInvalidMetricType
	}
	return nil
}

func (s *SeverUsecase) GetValue(metricType, metricName string) (string, error) {
	switch metricType {
	case "gauge":
		value, ok := s.repo.GetGauge(metricName)
		if !ok {
			return "", me.ErrMetricNotFound
		}
		str := float.FormatFloat(value)
		return str, nil
	case "counter":
		value, ok := s.repo.GetCounter(metricName)
		if !ok {
			return "", me.ErrMetricNotFound
		}
		return fmt.Sprintf("%d", value), nil
	default:
		return "", me.ErrInvalidMetricType
	}

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
		Gauges:   s.repo.GetAllGauge(),
		Counters: s.repo.GetAllCounter(),
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
