package storage

import (
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

// Restore Массово загружает переданные значения метрик в хранилища Storage.
func (s *Storage) Restore(m metrics.ProxyMetrics) error {
	// для удобства вызова PutMetrics проинициализируем нулевой хэш Gauges
	if m.Gauges == nil {
		m.Gauges = make(map[string]metrics.Gauge)
	}

	// для удобства вызова PutMetrics проинициализируем нулевой хэш Counters
	if m.Counters == nil {
		m.Counters = make(map[string]metrics.Counter)
	}

	s.gaugesMu.Lock()
	s.gauges = m.Gauges
	s.gaugesMu.Unlock()

	s.countersMu.Lock()
	s.counters = m.Counters
	s.countersMu.Unlock()

	return nil
}
