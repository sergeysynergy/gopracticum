package storage

import "github.com/sergeysynergy/metricser/pkg/metrics"

// GetMetrics Массово извлекает значение метрик из хранилища Storage.
func (s *Storage) GetMetrics() (*metrics.ProxyMetrics, error) {
	return s.repoDB.GetMetrics()
}
