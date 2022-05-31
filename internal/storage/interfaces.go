package storage

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

type Storer interface {
	Put(string, interface{}) error
	Get(string) (interface{}, error)

	PutMetrics(metrics.ProxyMetrics) error
	GetMetrics() (metrics.ProxyMetrics, error)

	Restore(metrics.ProxyMetrics) error
}

type DBStorer interface {
	Storer

	Ping() error
	Shutdown() error
}

type FileStorer interface {
	Storer
	WriteTicker() error
	WriteMetrics() (int, error)
	Shutdown() error
}
