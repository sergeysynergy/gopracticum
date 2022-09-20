package storage

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

type UseCase interface {
	Repo
	FileRepo

	SnapShotCreate() error
	WriteTicker() error
}

type Repo interface {
	Ping() error
	Shutdown() error

	Put(string, interface{}) error
	Get(string) (interface{}, error)

	PutMetrics(*metrics.ProxyMetrics) error
	GetMetrics() (*metrics.ProxyMetrics, error)

	Restore(*metrics.ProxyMetrics) error
}

type FileRepo interface {
	JustWriteMetrics(*metrics.ProxyMetrics) error
	JustReadMetrics() (*metrics.ProxyMetrics, error)
}
