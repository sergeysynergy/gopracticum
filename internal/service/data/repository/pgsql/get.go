package pgsql

import (
	"fmt"
	"github.com/sergeysynergy/metricser/internal/service/data/model"
	metricserErrors "github.com/sergeysynergy/metricser/internal/service/errors"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// Get извлекает из БД метрику любого типа по ID.
func (s *Storage) Get(id string) (interface{}, error) {
	m := model.Metrics{}
	row := s.db.QueryRowContext(
		s.ctx,
		`SELECT id, type, value, delta FROM metrics WHERE id=$1`,
		id,
	)
	// разбираем результат
	err := row.Scan(&m.ID, &m.MType, &m.Value, &m.Delta)
	if err != nil {
		return nil, err
	}

	switch m.MType {
	case metrics.TypeGauge:
		if !m.Value.Valid {
			return nil, fmt.Errorf("NULL gauge value")
		}
		return metrics.Gauge(m.Value.Float64), nil
	case metrics.TypeCounter:
		if !m.Delta.Valid {
			return nil, fmt.Errorf("NULL counter value")
		}
		return metrics.Counter(m.Delta.Int64), nil
	default:
	}

	return nil, metricserErrors.MetricNotImplemented
}
