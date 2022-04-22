package storage

import (
	"sync"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

type Storage struct {
	gaugesMu sync.RWMutex
	gauges   map[string]metrics.Gauge

	countersMu sync.RWMutex
	counters   map[string]metrics.Counter
}

type Options func(storage *Storage)

func New(opts ...Options) *Storage {
	s := &Storage{
		gauges:   make(map[string]metrics.Gauge, metrics.GaugeLen),
		counters: make(map[string]metrics.Counter, metrics.CounterLen),
	}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func WithGauges(gauges map[string]metrics.Gauge) Options {
	return func(s *Storage) {
		s.gauges = gauges
	}
}

func WithCounters(counters map[string]metrics.Counter) Options {
	return func(s *Storage) {
		s.counters = counters
	}
}

func (s *Storage) GetMetrics() metrics.ProxyMetric {
	return metrics.ProxyMetric{
		Gauges:   s.GetGauges(),
		Counters: s.GetCounters(),
	}
}
