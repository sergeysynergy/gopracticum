package storage

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

func (s *Storage) Put(key string, metric interface{}) error {
	switch m := metric.(type) {
	case metrics.Gauge:
		s.gaugesMu.Lock()
		s.gauges[key] = m
		s.gaugesMu.Unlock()
	case metrics.Counter:
		s.countersMu.Lock()
		_, ok := s.counters[key]
		if !ok {
			s.counters[key] = m
		} else {
			s.counters[key] += m
		}
		s.countersMu.Unlock()
	default:
		return ErrNotImplemented
	}

	return nil
}

func (s *Storage) PutMetrics(m metrics.ProxyMetrics) error {
	// для удобства вызова PutMetrics проиницилизируем нулевой хэш Gauges
	if m.Gauges == nil {
		m.Gauges = make(map[string]metrics.Gauge)
	}

	// для удобства вызова PutMetrics проиницилизируем нулевой хэш Counters
	if m.Counters == nil {
		m.Counters = make(map[string]metrics.Counter)
	}

	s.gaugesMu.Lock()
	for key, value := range m.Gauges {
		s.gauges[key] = value
	}
	s.gaugesMu.Unlock()

	s.countersMu.Lock()
	for key, delta := range m.Counters {
		_, ok := s.counters[key]
		if !ok {
			s.counters[key] = delta
		} else {
			s.counters[key] += delta
		}
	}
	s.countersMu.Unlock()

	return nil
}
