package memory

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"sync"
)

type Repo struct {
	gaugesMu sync.RWMutex
	gauges   map[string]metrics.Gauge

	countersMu sync.RWMutex
	counters   map[string]metrics.Counter
}

func New() *Repo {
	return &Repo{
		gauges:   make(map[string]metrics.Gauge, metrics.TypeGaugeLen),
		counters: make(map[string]metrics.Counter, metrics.TypeCounterLen),
	}
}
