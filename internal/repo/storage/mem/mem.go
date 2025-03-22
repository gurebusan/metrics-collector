package mem

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/zetcan333/metrics-collector/internal/models"
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

	existing, exists := s.Metrics[metric.ID]

	switch metric.MType {
	case models.Gauge:
		s.Metrics[metric.ID] = models.Metrics{
			MType: models.Gauge,
			ID:    metric.ID,
			Value: metric.Value,
		}
	case models.Counter:
		var newDelta int64
		if exists && existing.Delta != nil {
			newDelta = *existing.Delta
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

func (s *MemStorage) GetMetric(id string) (models.Metrics, bool) {
	s.RLock()
	defer s.RUnlock()
	metric, exists := s.Metrics[id]
	return metric, exists
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

func (s *MemStorage) SaveToFile() error {
	s.RLock()
	defer s.RUnlock()

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(wd, "../../backup/metrics-db.json")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, metric := range s.Metrics {
		if err := encoder.Encode(metric); err != nil {
			return err
		}
	}
	return nil

}

func (s *MemStorage) LoadFromFile() error {
	s.Lock()
	defer s.Unlock()

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(wd, "../../backup/metrics-db.json")
	file, err := os.Open(path)

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	s.Metrics = make(map[string]models.Metrics)

	for scanner.Scan() {
		var metric models.Metrics
		err := json.Unmarshal(scanner.Bytes(), &metric)
		if err != nil {
			return fmt.Errorf("ошибка декодирования JSON: %v", err)
		}
		s.Metrics[metric.ID] = metric
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ошибка чтения файла: %v", err)
	}

	return nil
}
