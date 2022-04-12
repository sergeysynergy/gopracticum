package storage

import "github.com/sergeysynergy/gopracticum/pkg/metrics"

type Storer interface {
	PutGauge(string, metrics.Gauge)
	GetGauge(string) (metrics.Gauge, error)
	GetGauges() map[string]metrics.Gauge
	BulkPutGauges(map[string]metrics.Gauge)

	IncreaseCounter(string)
	PostCounter(string, metrics.Counter)
	GetCounter(string) (metrics.Counter, error)
	GetCounters() map[string]metrics.Counter
	BulkPutCounters(map[string]metrics.Counter)
}
