package storage

import (
	"fmt"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

// IncreaseCounter увеличиваем значение счётчика на единицу
func (s *Storage) IncreaseCounter(key string) {
	s.countersMu.Lock()
	defer s.countersMu.Unlock()

	_, ok := s.counters[key]
	if !ok {
		s.counters[key] = 1
		return
	}

	s.counters[key] += 1
}

// PostCounter прибавляем к текущему значению счётчика переданное значение
func (s *Storage) PostCounter(key string, val metrics.Counter) {
	s.countersMu.Lock()
	defer s.countersMu.Unlock()

	_, ok := s.counters[key]
	if !ok {
		s.counters[key] = val
		return
	}

	s.counters[key] += val
}

func (s *Storage) GetCounter(key string) (metrics.Counter, error) {
	s.countersMu.RLock()
	defer s.countersMu.RUnlock()

	counter, ok := s.counters[key]
	if !ok {
		return 0, fmt.Errorf("counter metric with key '%s' not found", key)
	}

	return counter, nil
}

// GetCounters получаем значение всех метрик counter
func (s *Storage) GetCounters() map[string]metrics.Counter {
	s.countersMu.RLock()
	defer s.countersMu.RUnlock()

	return s.counters
}

// BulkPutCounters массово перезаписываем значения всех метрик counter
func (s *Storage) BulkPutCounters(counters map[string]metrics.Counter) {
	s.countersMu.Lock()
	defer s.countersMu.Unlock()

	s.counters = counters
}
