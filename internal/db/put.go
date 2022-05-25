package db

import (
	"database/sql"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
	"log"
)

func (s *Storage) Put(id string, val interface{}) error {
	switch m := val.(type) {
	case metrics.Gauge:
		result, err := s.db.ExecContext(s.ctx, queryUpdateGauge, id, m)
		if err != nil {
			return err
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			_, err = s.db.ExecContext(s.ctx, queryInsertGauge, id, m)
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
		return metrics.ErrNotImplemented
	}

	return nil
}

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
				log.Println("db.PutMetrics CPUutilization1:", value)
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
			mdb := metricsDB{}
			// s.db.PrepareContext(s.ctx, "SELECT id, type, value, delta FROM metrics WHERE id=$1")
			row := txCounterGet.QueryRowContext(s.ctx, id)
			err = row.Scan(&mdb.ID, &mdb.MType, &mdb.Value, &mdb.Delta)
			if err == sql.ErrNoRows {
				// добавим новую запись в случае отсутствия результата
				// s.db.PrepareContext(s.ctx, "INSERT INTO metrics (id, type, delta) VALUES ($1, 'counter', $2)")
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
			if mdb.Delta.Valid {
				v = metrics.Counter(mdb.Delta.Int64)
			}
			hm := delta + v
			// s.db.PrepareContext(s.ctx, "UPDATE metrics SET delta = $2 WHERE id = $1")
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
