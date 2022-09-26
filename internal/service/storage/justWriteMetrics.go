package storage

import "github.com/sergeysynergy/metricser/pkg/metrics"

func (s *Storage) JustWriteMetrics(prm *metrics.ProxyMetrics) (err error) {
	if s.fileRepo != nil {
		err = s.fileRepo.JustWriteMetrics(prm)
		if err != nil {
			return err
		}
	}

	return nil
}
