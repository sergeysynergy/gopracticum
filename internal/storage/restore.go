package storage

import (
	metricserErrors "github.com/sergeysynergy/metricser/internal/errors"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// Restore Массово загружает переданные значения метрик в репозиторий.
func (s *Storage) Restore(prm *metrics.ProxyMetrics) error {
	if prm == nil {
		return metricserErrors.ErrEmptyFilename
	}

	return s.repo.Restore(prm)
}
