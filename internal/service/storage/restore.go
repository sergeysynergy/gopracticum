package storage

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// Restore Массово загружает переданные значения метрик в хранилища Storage.
func (s *Storage) Restore(prm *metrics.ProxyMetrics) error {
	return s.repo.Restore(prm)
}
