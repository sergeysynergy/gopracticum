package storage

import (
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

type Repo interface {
	Put(string, interface{}) error
	Get(string) (interface{}, error)

	PutMetrics(metrics.ProxyMetrics) error
	GetMetrics() (metrics.ProxyMetrics, error)

	Restore(metrics.ProxyMetrics) error
}

type DBStorer interface {
	Repo

	Ping() error
	Shutdown() error
}

type FileStorer interface {
	Repo
	WriteTicker() error
	WriteMetrics() (int, error)
	Shutdown() error
}
