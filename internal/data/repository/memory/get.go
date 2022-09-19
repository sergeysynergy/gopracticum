package memory

import (
	"fmt"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// Get Извлекает значение метрики из хранилища Storage для заданного ID.
func (r *Repo) Get(id string) (interface{}, error) {
	r.countersMu.Lock()
	defer r.countersMu.Unlock()
	delta, ok := r.counters[id]
	if ok {
		return delta, nil
	}

	r.gaugesMu.Lock()
	defer r.gaugesMu.Unlock()
	value, ok := r.gauges[id]
	if ok {
		return value, nil
	}

	return nil, fmt.Errorf("metrics not found")
}

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
