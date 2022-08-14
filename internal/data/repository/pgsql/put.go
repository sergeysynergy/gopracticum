package pgsql

import (
	"database/sql"
	"github.com/sergeysynergy/gopracticum/internal/data/model"
	metricserErrors "github.com/sergeysynergy/gopracticum/internal/errors"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
	"log"
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
		// запишим увеличенное значение
		if _, err = s.stmtCounterUpdate.ExecContext(s.ctx, id, m+v); err != nil {
			return err
		}
	default:
		return metricserErrors.MetricNotImplemented
	}

	return nil
}

// PutMetrics Массово записывает значение метрик в БД.
func (s *Storage) PutMetrics(m metrics.ProxyMetrics) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txGaugeUpdate := tx.StmtContext(s.ctx, s.stmtGaugeUpdate)
	txCounterUpdate := tx.StmtContext(s.ctx, s.stmtCounterUpdate)
	txGaugeInsert := tx.StmtContext(s.ctx, s.stmtGaugeInsert)
	txCounterInsert := tx.StmtContext(s.ctx, s.stmtCounterInsert)
	txCounterGet := tx.StmtContext(s.ctx, s.stmtCounterGet)

	if m.Gauges != nil {
		for id, value := range m.Gauges {
			if id == "CPUutilization1" {
				log.Println("pgsql.PutMetrics CPUutilization1:", value)
			}

			result, err := txGaugeUpdate.Exec(id, value)
			if err != nil {
				return err
			}
			count, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if count == 0 {
				_, err := txGaugeInsert.ExecContext(s.ctx, id, value)
				if err != nil {
					return err
				}
			}
		}
	}

	if m.Counters != nil {
		for id, delta := range m.Counters {
			// получим текущее значение счётчика
			m := model.Metrics{}
			// s.pgsql.PrepareContext(s.ctx, "SELECT id, type, value, delta FROM metrics WHERE id=$1")
			row := txCounterGet.QueryRowContext(s.ctx, id)
			err = row.Scan(&m.ID, &m.MType, &m.Value, &m.Delta)
			if err == sql.ErrNoRows {
				// добавим новую запись в случае отсутствия результата
				// s.pgsql.PrepareContext(s.ctx, "INSERT INTO metrics (id, type, delta) VALUES ($1, 'counter', $2)")
				_, err = txCounterInsert.ExecContext(s.ctx, id, delta)
				if err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}

			// запишим увеличенное значение
			v := metrics.Counter(0)
			if m.Delta.Valid {
				v = metrics.Counter(m.Delta.Int64)
			}
			hm := delta + v
			// s.pgsql.PrepareContext(s.ctx, "UPDATE metrics SET delta = $2 WHERE id = $1")
			if _, err = txCounterUpdate.ExecContext(s.ctx, id, hm); err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[ERROR] put metrics transaction failed - ", err)
		return err
	}

	return nil
}