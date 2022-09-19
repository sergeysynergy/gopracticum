package memory

import "github.com/sergeysynergy/metricser/pkg/metrics"

// PutMetrics Массово записывает значение метрик в хранилище Storage.
func (r *Repo) PutMetrics(m *metrics.ProxyMetrics) error {
	// для удобства вызова PutMetrics проинициализируем нулевой хэш Gauges
	if m.Gauges == nil {
		m.Gauges = make(map[string]metrics.Gauge)
	}

	// для удобства вызова PutMetrics проинициализируем нулевой хэш Counters
	if m.Counters == nil {
		m.Counters = make(map[string]metrics.Counter)
	}

	r.gaugesMu.Lock()
	for key, value := range m.Gauges {
		r.gauges[key] = value
	}
	r.gaugesMu.Unlock()

	r.countersMu.Lock()
	for key, delta := range m.Counters {
		_, ok := r.counters[key]
		if !ok {
			r.counters[key] = delta
		} else {
			r.counters[key] += delta
		}
	}
	r.countersMu.Unlock()

	return nil
}
