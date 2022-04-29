package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
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

func (s *Storage) Put(_ context.Context, key string, metric interface{}) error {
	switch m := metric.(type) {
	case metrics.Gauge:
		s.gaugesMu.Lock()
		defer s.gaugesMu.Unlock()
		s.gauges[key] = m
	case metrics.Counter:
		s.countersMu.Lock()
		defer s.countersMu.Unlock()

		_, ok := s.counters[key]
		if !ok {
			s.counters[key] = m
		} else {
			s.counters[key] += m
		}
	default:
		return ErrNotImplemented
	}

	return nil
}

func (s *Storage) Get(_ context.Context, key string) (interface{}, error) {
	s.countersMu.Lock()
	defer s.countersMu.Unlock()
	delta, ok := s.counters[key]
	if ok {
		return delta, nil
	}

	s.gaugesMu.Lock()
	defer s.gaugesMu.Unlock()
	value, ok := s.gauges[key]
	if ok {
		return value, nil
	}

	return nil, ErrNotFound
}

func (s *Storage) PutMetrics(_ context.Context, m metrics.ProxyMetrics) error {
	// для удобства вызова PutMetrics проиницилизируем нулевой хэш Gauges
	if m.Gauges == nil {
		m.Gauges = make(map[string]metrics.Gauge)
	}

	// для удобства вызова PutMetrics проиницилизируем нулевой хэш Counters
	if m.Counters == nil {
		m.Counters = make(map[string]metrics.Counter)
	}

	s.gaugesMu.Lock()
	s.gauges = m.Gauges
	s.gaugesMu.Unlock()

	s.countersMu.Lock()
	s.counters = m.Counters
	s.countersMu.Unlock()

	return nil
}

func (s *Storage) GetMetrics(_ context.Context) (metrics.ProxyMetrics, error) {
	s.gaugesMu.RLock()
	gauges := s.gauges
	s.gaugesMu.RUnlock()

	s.countersMu.RLock()
	counters := s.counters
	s.countersMu.RUnlock()

	return metrics.ProxyMetrics{
		Gauges:   gauges,
		Counters: counters,
	}, nil
}
