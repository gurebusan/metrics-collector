package mem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/zetcan333/metrics-collector/internal/models"
	"github.com/zetcan333/metrics-collector/pkg/myerrors"
)

type MemStorage struct {
	sync.RWMutex
	Metrics map[string]models.Metrics
}

func NewStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]models.Metrics),
	}
}
func (s *MemStorage) UpdateMetric(ctx context.Context, metric models.Metrics) error {
	s.Lock()
	defer s.Unlock()

	currentMetric, exists := s.Metrics[metric.ID]

	switch metric.MType {
	case models.Gauge:
		s.Metrics[metric.ID] = models.Metrics{
			MType: models.Gauge,
			ID:    metric.ID,
			Value: metric.Value,
		}
	case models.Counter:
		var newDelta int64
		if exists && currentMetric.Delta != nil {
			newDelta = *currentMetric.Delta
		}
		if metric.Delta != nil {
			newDelta += *metric.Delta
		}
		s.Metrics[metric.ID] = models.Metrics{
			MType: models.Counter,
			ID:    metric.ID,
			Delta: &newDelta,
		}
	}
	return nil
}

func (s *MemStorage) GetMetric(ctx context.Context, id string) (models.Metrics, error) {
	s.RLock()
	defer s.RUnlock()
	metric, exists := s.Metrics[id]
	if !exists {
		return models.Metrics{}, myerrors.ErrMetricNotFound
	}
	return metric, nil
}

func (s *MemStorage) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	s.RLock()
	defer s.RUnlock()
	all := make(map[string]float64)
	for key, value := range s.Metrics {
		if value.Value != nil {
			all[key] = *value.Value
		}
	}
	return all, nil
}

func (s *MemStorage) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	s.RLock()
	defer s.RUnlock()
	all := make(map[string]int64)
	for key, value := range s.Metrics {
		if value.Delta != nil {
			all[key] = *value.Delta
		}
	}
	return all, nil
}

func (s *MemStorage) UpdateMetricsWithBatch(ctx context.Context, metrics []models.Metrics) error {
	s.Lock()
	defer s.Unlock()

	for _, metric := range metrics {
		s.updateMetricUnsafe(metric)
	}
	return nil
}

func (s *MemStorage) updateMetricUnsafe(metric models.Metrics) {
	currentMetric, exists := s.Metrics[metric.ID]

	switch metric.MType {
	case models.Gauge:
		s.Metrics[metric.ID] = models.Metrics{
			MType: models.Gauge,
			ID:    metric.ID,
			Value: metric.Value,
		}
	case models.Counter:
		var newDelta int64
		if exists && currentMetric.Delta != nil {
			newDelta = *currentMetric.Delta
		}
		if metric.Delta != nil {
			newDelta += *metric.Delta
		}
		s.Metrics[metric.ID] = models.Metrics{
			MType: models.Counter,
			ID:    metric.ID,
			Delta: &newDelta,
		}
	}
}

func (s *MemStorage) SaveBkpToFile(path string) error {
	const op = "internal.repo.storage.mem.SaveBkpToFile"
	s.RLock()
	defer s.RUnlock()

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer file.Close()

	// Сериализуем всю мапу одним JSON-объектом
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(s.Metrics); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *MemStorage) LoadBkpFromFile(path string) error {
	const op = "internal.repo.storage.mem.LoadBkpFromFile"
	s.Lock()
	defer s.Unlock()

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	defer file.Close()

	// Декодируем весь JSON-файл в мапу
	var metrics map[string]models.Metrics
	if err := json.NewDecoder(file).Decode(&metrics); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	s.Metrics = metrics
	return nil
}

// mock Ping
func (s *MemStorage) Ping(ctx context.Context) error {
	return nil
}

// mock InitTable
func (s *MemStorage) InitTable(ctx context.Context) error {
	return nil
}
