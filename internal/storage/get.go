package storage

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

func (s *Storage) Get(key string) (interface{}, error) {
	s.countersMu.Lock()
	defer s.countersMu.Unlock()
	delta, ok := s.counters[key]
	if ok {
		return delta, nil
	}

	s.gaugesMu.Lock()
	defer s.gaugesMu.Unlock()
	value, ok := s.gauges[key]
	if ok {
		return value, nil
	}

	return nil, ErrNotFound
}

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
