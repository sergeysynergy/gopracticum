package storage

import (
	"fmt"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

func (s *Storage) WriteTicker(prm *metrics.ProxyMetrics) error {
	if s.fileRepo == nil {
		return fmt.Errorf("empty filestore repository")
	}

	prm, err := s.repo.GetMetrics()
	if err != nil {
		return err
	}

	return s.fileRepo.WriteTicker(prm)
}
