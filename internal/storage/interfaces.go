package storage

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

type UseCase interface {
	Repo
	FileRepo
}

type Repo interface {
	Shutdown() error
	PutMetrics(*metrics.ProxyMetrics) error
	GetMetrics() (*metrics.ProxyMetrics, error)
	Restore(*metrics.ProxyMetrics) error
	Ping() error
	Put(string, interface{}) error
	Get(string) (interface{}, error)
}

type FileRepo interface {
	WriteMetrics(*metrics.ProxyMetrics) error
	ReadMetrics() (*metrics.ProxyMetrics, error)

	WriteTicker(*metrics.ProxyMetrics) error
}
