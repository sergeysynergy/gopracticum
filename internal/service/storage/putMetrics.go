package storage

import (
	serviceErrors "github.com/sergeysynergy/metricser/internal/service/errors"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// PutMetrics Массово записывает значение метрик в хранилище Storage.
func (s *Storage) PutMetrics(prm *metrics.ProxyMetrics) error {
	if len(prm.Gauges) == 0 && len(prm.Counters) == 0 {
		return serviceErrors.ErrEmptyProxyMetrics
	}

	return s.repo.PutMetrics(prm)
}
