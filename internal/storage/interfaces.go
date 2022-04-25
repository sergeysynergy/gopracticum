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

type RepoStorer interface {
	Storer
	WriteTicker() error
	WriteMetrics() (int, error)
	Shutdown() error
}
