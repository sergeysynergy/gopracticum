package handlers

import "github.com/sergeysynergy/gopracticum/pkg/metrics"

type Storer interface {
	PutGauge(string, metrics.Gauge)
	GetGauge(string) (metrics.Gauge, error)
	Gauges() map[string]metrics.Gauge
	PostCounter(string, metrics.Counter)
	GetCounter(string) (metrics.Counter, error)
	Counters() map[string]metrics.Counter
}
