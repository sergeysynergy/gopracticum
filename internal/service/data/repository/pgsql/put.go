package pgsql

import (
	metricserErrors "github.com/sergeysynergy/metricser/internal/service/errors"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// Put записывает значение метрики в БД для заданного ID.
func (s *Storage) Put(id string, val interface{}) error {
	switch m := val.(type) {
	case metrics.Gauge:
		result, err := s.db.ExecContext(s.ctx, `UPDATE metrics SET value = $2 WHERE id = $1`, id, m)
		if err != nil {
			return err
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			_, err = s.db.ExecContext(
				s.ctx,
				`INSERT INTO metrics (id, type, value) VALUES ($1, 'gauge', $2)`,
				id,
				m,
			)
			if err != nil {
				return err
			}
		}
	case metrics.Counter:
		// получим текущее значение счётчика
		v, err := s.checkCounter(id)
		if err != nil {
			return err
		}
		// запишем увеличенное значение
		if _, err = s.stmtCounterUpdate.ExecContext(s.ctx, id, m+v); err != nil {
			return err
		}
	default:
		return metricserErrors.MetricNotImplemented
	}

	return nil
}
