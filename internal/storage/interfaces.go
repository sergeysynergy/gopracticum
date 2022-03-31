package storage

import "github.com/sergeysynergy/gopracticum/pkg/metrics"

type Repositories interface {
	Put(metrics.Name, metrics.Gauge)
	Count(metrics.Name, metrics.Counter)
}
