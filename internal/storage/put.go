package storage

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// Put Записывает значение метрики в хранилище Storage для заданного ID.
func (s *Storage) Put(id string, metric interface{}) error {
	return s.repo.Put(id, metric)
}

// PutMetrics Массово записывает значение метрик в хранилище Storage.
func (s *Storage) PutMetrics(m metrics.ProxyMetrics) error {
	return s.repo.PutMetrics(m)
}
