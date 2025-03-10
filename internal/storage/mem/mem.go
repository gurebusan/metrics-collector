package mem

import "sync"

type MemStorage struct {
	sync.RWMutex
	gauge   map[string]float64
	counter map[string]int64
}

func NewStorage() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: map[string]int64{},
	}
}
func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.Lock()
	defer s.Unlock()
	s.gauge[name] = value
}

func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.Lock()
	defer s.Unlock()
	s.counter[name] += value
}

func (s *MemStorage) GetGauge(name string) (float64, bool) {
	s.RLock()
	defer s.RUnlock()
	val, ok := s.gauge[name]
	return val, ok
}

func (s *MemStorage) GetCounter(name string) (int64, bool) {
	s.RLock()
	defer s.RUnlock()
	val, ok := s.counter[name]
	return val, ok
}
