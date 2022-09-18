package storage

import (
	"fmt"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

func (s *Storage) WriteMetrics(prm *metrics.ProxyMetrics) error {
	if s.repoFile == nil {
		return fmt.Errorf("file store not defined")
	}

	return s.repoFile.WriteMetrics(prm)
}
