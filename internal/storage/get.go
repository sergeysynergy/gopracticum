package storage

import (
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

// Get Извлекает значение метрики из хранилища Storage для заданного ID.
func (s *Storage) Get(id string) (interface{}, error) {
	s.countersMu.Lock()
	defer s.countersMu.Unlock()
	delta, ok := s.counters[id]
	if ok {
		return delta, nil
	}

	s.gaugesMu.Lock()
	defer s.gaugesMu.Unlock()
	value, ok := s.gauges[id]
	if ok {
		return value, nil
	}

	return nil, ErrNotFound
}

// GetMetrics Массово извлекает значение метрик из хранилища Storage.
func (s *Storage) GetMetrics() (metrics.ProxyMetrics, error) {
	prm := metrics.NewProxyMetrics()

	s.gaugesMu.RLock()
	for k, v := range s.gauges {
		prm.Gauges[k] = v

	}
	s.gaugesMu.RUnlock()

	s.countersMu.RLock()
	for k, v := range s.counters {
		prm.Counters[k] = v

	}
	s.countersMu.RUnlock()

	return prm, nil
}
