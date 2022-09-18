package storage

import "github.com/sergeysynergy/metricser/pkg/metrics"

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Put interface {
	Put(string, interface{}) error
}

type Get interface {
	Get(string) (interface{}, error)
}

type PutMetrics interface {
	PutMetrics(*metrics.ProxyMetrics) error
}

type GetMetrics interface {
	GetMetrics() (*metrics.ProxyMetrics, error)
}

type Ping interface {
	Ping() error
}

type Shutdown interface {
	Shutdown() error
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type RepoDB interface {
	Ping
	Put
	Get
	GetMetrics
	PutMetrics
	Shutdown
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ReadMetrics interface {
	ReadMetrics() (*metrics.ProxyMetrics, error)
}

type WriteMetrics interface {
	WriteMetrics(*metrics.ProxyMetrics) error
}

type RepoFile interface {
	ReadMetrics
	WriteMetrics
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type WriteTicker interface {
	WriteTicker() error
}

type UseCase interface {
	WriteTicker
	Shutdown

	Put
	Get
	PutMetrics
	GetMetrics
	Ping
	WriteMetrics
}
