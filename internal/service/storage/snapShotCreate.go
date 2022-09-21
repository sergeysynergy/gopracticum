package storage

import (
	serviceErrors "github.com/sergeysynergy/metricser/internal/service/errors"
)

func (s *Storage) SnapShotCreate() (err error) {
	if s.fileRepo == nil {
		return serviceErrors.ErrFileStoreNotDefined
	}

	prm, err := s.repo.GetMetrics()
	if err != nil {
		return err
	}

	return s.fileRepo.JustWriteMetrics(prm)
}
