package storage

import (
	"fmt"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

func (s *Storage) WriteMetrics(prm *metrics.ProxyMetrics) error {
	if s.fileRepo == nil {
		return fmt.Errorf("empty filestore repository")
	}

	return s.fileRepo.WriteMetrics(prm)
}
