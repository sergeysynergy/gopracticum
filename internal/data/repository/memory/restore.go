package memory

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// Restore Массово загружает переданные значения метрик в хранилища Storage.
func (r *Repo) Restore(m metrics.ProxyMetrics) error {
	if m.Gauges == nil {
		m.Gauges = make(map[string]metrics.Gauge)
	}

	if m.Counters == nil {
		m.Counters = make(map[string]metrics.Counter)
	}

	r.gaugesMu.Lock()
	r.gauges = m.Gauges
	r.gaugesMu.Unlock()

	r.countersMu.Lock()
	r.counters = m.Counters
	r.countersMu.Unlock()

	return nil
}
