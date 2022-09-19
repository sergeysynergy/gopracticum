package storage

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"

	metricserErrors "github.com/sergeysynergy/metricser/internal/errors"
)

// ReadMetrics Массово извлекает значение метрик из хранилища Storage.
func (s *Storage) ReadMetrics() (*metrics.ProxyMetrics, error) {
	if s.fileRepo != nil {

		return s.fileRepo.ReadMetrics()
	}
	return nil, metricserErrors.ErrEmptyFilename
}
