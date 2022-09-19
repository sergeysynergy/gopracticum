package memory

import (
	"fmt"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// Put Записывает значение метрики в хранилище Storage для заданного ID.
func (r *Repo) Put(id string, metric interface{}) error {
	switch m := metric.(type) {
	case metrics.Gauge:
		r.gaugesMu.Lock()
		r.gauges[id] = m
		r.gaugesMu.Unlock()
	case metrics.Counter:
		r.countersMu.Lock()
		_, ok := r.counters[id]
		if !ok {
			r.counters[id] = m
		} else {
			r.counters[id] += m
		}
		r.countersMu.Unlock()
	default:
		return fmt.Errorf("metrics not implemented")
	}

	return nil
}
