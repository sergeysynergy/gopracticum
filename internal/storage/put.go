package storage

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// Put Записывает значение метрики в хранилище Storage для заданного ID.
func (s *Storage) Put(id string, metric interface{}) error {
	switch m := metric.(type) {
	case metrics.Gauge:
		s.gaugesMu.Lock()
		s.gauges[id] = m
		s.gaugesMu.Unlock()
	case metrics.Counter:
		s.countersMu.Lock()
		_, ok := s.counters[id]
		if !ok {
			s.counters[id] = m
		} else {
			s.counters[id] += m
		}
		s.countersMu.Unlock()
	default:
		return ErrNotImplemented
	}

	return nil
}

// PutMetrics Массово записывает значение метрик в хранилище Storage.
func (s *Storage) PutMetrics(m metrics.ProxyMetrics) error {
	// для удобства вызова PutMetrics проинициализируем нулевой хэш Gauges
	if m.Gauges == nil {
		m.Gauges = make(map[string]metrics.Gauge)
	}

	// для удобства вызова PutMetrics проинициализируем нулевой хэш Counters
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
