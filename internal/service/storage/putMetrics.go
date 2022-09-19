package storage

import "github.com/sergeysynergy/metricser/pkg/metrics"

// PutMetrics Массово записывает значение метрик в хранилище Storage.
func (s *Storage) PutMetrics(prm *metrics.ProxyMetrics) error {
	return s.repo.PutMetrics(prm)
}
