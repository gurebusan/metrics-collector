package mem

import (
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
func (s *MemStorage) UpdateMetric(metric models.Metrics) {
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
}

func (s *MemStorage) GetMetric(id string) (models.Metrics, error) {
	s.RLock()
	defer s.RUnlock()
	metric, exists := s.Metrics[id]
	if !exists {
		return models.Metrics{}, myerrors.ErrMetricNotFound
	}
	return metric, nil
}

func (s *MemStorage) GetAllGauges() map[string]float64 {
	s.RLock()
	defer s.RUnlock()
	all := make(map[string]float64)
	for key, value := range s.Metrics {
		if value.Value != nil {
			all[key] = *value.Value
		}
	}
	return all
}

func (s *MemStorage) GetAllCounters() map[string]int64 {
	s.RLock()
	defer s.RUnlock()
	all := make(map[string]int64)
	for key, value := range s.Metrics {
		if value.Delta != nil {
			all[key] = *value.Delta
		}
	}
	return all
}

func (s *MemStorage) SaveBkpToFile(path string) error {
	s.RLock()
	defer s.RUnlock()

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Сериализуем всю мапу одним JSON-объектом
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(s.Metrics); err != nil {
		return fmt.Errorf("ошибка при сериализации данных: %v", err)
	}

	return nil
}

func (s *MemStorage) LoadBkpFromFile(path string) error {
	s.Lock()
	defer s.Unlock()

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	// Декодируем весь JSON-файл в мапу
	var metrics map[string]models.Metrics
	if err := json.NewDecoder(file).Decode(&metrics); err != nil {
		return fmt.Errorf("ошибка при десериализации данных: %v", err)
	}

	s.Metrics = metrics
	return nil
}
