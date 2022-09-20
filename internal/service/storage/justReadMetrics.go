package storage

import (
	serviceErrors "github.com/sergeysynergy/metricser/internal/service/errors"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

func (s *Storage) JustReadMetrics() (prm *metrics.ProxyMetrics, err error) {
	if s.fileRepo == nil {
		return nil, serviceErrors.ErrFileStoreNotDefined
	}

	prm, err = s.fileRepo.JustReadMetrics()
	if err != nil {
		return nil, err
	}

	return prm, nil
}
