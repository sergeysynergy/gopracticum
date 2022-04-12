package storage

import (
	"fmt"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

func (s *Storage) PutGauge(key string, val metrics.Gauge) {
	s.gaugesMu.Lock()
	defer s.gaugesMu.Unlock()

	s.gauges[key] = val
}

func (s *Storage) GetGauge(key string) (metrics.Gauge, error) {
	s.gaugesMu.RLock()
	defer s.gaugesMu.RUnlock()

	gauge, ok := s.gauges[key]
	if !ok {
		return 0, fmt.Errorf("gauge metric with key '%s' not found", key)
	}

	return gauge, nil
}

// GetGauges получаем значение всех метрик gauge
func (s *Storage) GetGauges() map[string]metrics.Gauge {
	s.gaugesMu.RLock()
	defer s.gaugesMu.RUnlock()

	return s.gauges
}

// BulkPutGauges массово перезаписываем значения всех метрик gauge
func (s *Storage) BulkPutGauges(gauges map[string]metrics.Gauge) {
	s.gaugesMu.Lock()
	defer s.gaugesMu.Unlock()

	s.gauges = gauges
}
