package storage

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

func (s *Storage) PutMetrics(prm *metrics.ProxyMetrics) error {
	return s.repo.PutMetrics(prm)
}
