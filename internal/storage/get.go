package storage

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// Get Извлекает значение метрики из хранилища Storage для заданного ID.
func (s *Storage) Get(id string) (interface{}, error) {
	return s.repo.Get(id)
}

// GetMetrics Массово извлекает значение метрик из хранилища Storage.
func (s *Storage) GetMetrics() (metrics.ProxyMetrics, error) {
	return s.repo.GetMetrics()
}
