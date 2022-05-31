package storage

import (
	"errors"
	"sync"

	"github.com/sergeysynergy/metricser/pkg/metrics"
)

var (
	ErrNotImplemented = errors.New("storage: metric not implemented")
	ErrNotFound       = errors.New("storage: metric not found")
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
		gauges:   make(map[string]metrics.Gauge, metrics.TypeGaugeLen),
		counters: make(map[string]metrics.Counter, metrics.TypeCounterLen),
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
