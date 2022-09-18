package memory

import "github.com/sergeysynergy/metricser/pkg/metrics"

func (r *Repo) PutMetrics(prm *metrics.ProxyMetrics) error {
	if prm.Gauges != nil {
		r.gaugesMu.Lock()
		for key, value := range prm.Gauges {
			r.gauges[key] = value
		}
		r.gaugesMu.Unlock()
	}

	if prm.Counters != nil {
		r.countersMu.Lock()
		for key, delta := range prm.Counters {
			_, ok := r.counters[key]
			if !ok {
				r.counters[key] = delta
			} else {
				r.counters[key] += delta
			}
		}
		r.countersMu.Unlock()
	}

	return nil
}
