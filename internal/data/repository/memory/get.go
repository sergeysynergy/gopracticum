package memory

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

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

	return nil, metrics.ErrNotFound
}
