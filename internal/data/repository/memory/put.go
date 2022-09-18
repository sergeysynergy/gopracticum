package memory

import "github.com/sergeysynergy/metricser/pkg/metrics"

func (r *Repo) Put(id string, mrc interface{}) error {
	switch m := mrc.(type) {
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
		return metrics.ErrNotImplemented
	}

	return nil
}
