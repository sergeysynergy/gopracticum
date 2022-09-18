package storage

import "github.com/sergeysynergy/metricser/pkg/metrics"

// PutMetrics Массово записывает значение метрик в хранилище Storage.
func (s *Storage) PutMetrics(prm *metrics.ProxyMetrics) error {
	err := s.repoDB.PutMetrics(prm)
	if err != nil {
		return err
	}
	return nil

}
