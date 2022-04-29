package storage

import (
	"context"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

type Storer interface {
	Put(context.Context, string, interface{}) error
	Get(context.Context, string) (interface{}, error)

	PutMetrics(context.Context, metrics.ProxyMetrics) error
	GetMetrics(context.Context) (metrics.ProxyMetrics, error)
}

type DBStorer interface {
	Ping() error
	Shutdown() error

	Put(context.Context, string, interface{}) error
	Get(context.Context, string) (interface{}, error)

	PutMetrics(context.Context, metrics.ProxyMetrics) error
	GetMetrics(context.Context) (metrics.ProxyMetrics, error)
}

type FileStorer interface {
	Storer
	WriteTicker() error
	WriteMetrics() (int, error)
	Shutdown() error
}
