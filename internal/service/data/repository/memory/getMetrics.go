package memory

import "github.com/sergeysynergy/metricser/pkg/metrics"

// GetMetrics Массово извлекает значение метрик из хранилища Storage.
func (r *Repo) GetMetrics() (*metrics.ProxyMetrics, error) {
	prm := metrics.NewProxyMetrics()

	r.gaugesMu.RLock()
	for k, v := range r.gauges {
		prm.Gauges[k] = v

	}
	r.gaugesMu.RUnlock()

	r.countersMu.RLock()
	for k, v := range r.counters {
		prm.Counters[k] = v

	}
	r.countersMu.RUnlock()

	return prm, nil
}
