package storage

import (
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

type Storer interface {
	Put(string, interface{}) error
	Get(string) (interface{}, error)

	PutMetrics(metrics.ProxyMetric)
	GetMetrics() metrics.ProxyMetric
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
