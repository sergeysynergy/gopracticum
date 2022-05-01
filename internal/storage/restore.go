package storage

import (
	"context"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

func (s *Storage) Restore(_ context.Context, m metrics.ProxyMetrics) error {
	// для удобства вызова PutMetrics проиницилизируем нулевой хэш Gauges
	if m.Gauges == nil {
		m.Gauges = make(map[string]metrics.Gauge)
	}

	// для удобства вызова PutMetrics проиницилизируем нулевой хэш Counters
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
