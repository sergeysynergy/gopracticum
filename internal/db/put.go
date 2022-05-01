package db

import (
	"context"
	"database/sql"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
	"log"
)

func (s *Storage) Put(parentCtx context.Context, id string, val interface{}) error {
	ctx, cancel := context.WithTimeout(parentCtx, queryTimeOut)
	defer cancel()

	switch m := val.(type) {
	case metrics.Gauge:
		result, err := s.db.ExecContext(ctx, queryUpdateGauge, id, m)
		if err != nil {
			return err
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			_, err = s.db.ExecContext(ctx, queryInsertGauge, id, m)
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
		if _, err = s.stmtCounterUpdate.ExecContext(ctx, id, m+v); err != nil {
			return err
		}
	default:
		return metrics.ErrNotImplemented
	}

	return nil
}

func (s *Storage) PutMetrics(ctx context.Context, m metrics.ProxyMetrics) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txGaugeUpdate := tx.StmtContext(ctx, s.stmtGaugeUpdate)
	txCounterUpdate := tx.StmtContext(ctx, s.stmtCounterUpdate)
	txGaugeInsert := tx.StmtContext(ctx, s.stmtGaugeInsert)
	txCounterInsert := tx.StmtContext(ctx, s.stmtCounterInsert)
	txCounterGet := tx.StmtContext(ctx, s.stmtCounterGet)

	if m.Gauges != nil {
		for id, value := range m.Gauges {
			result, err := txGaugeUpdate.Exec(id, value)
			if err != nil {
				return err
			}
			count, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if count == 0 {
				_, err := txGaugeInsert.ExecContext(ctx, id, value)
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
			row := txCounterGet.QueryRowContext(s.ctx, id)
			err = row.Scan(&mdb.ID, &mdb.MType, &mdb.Value, &mdb.Delta)
			if err == sql.ErrNoRows {
				_, err = txCounterInsert.ExecContext(s.ctx, id, delta)
				if err != nil {
					return err
				}
			}
			if err != nil {
				return err
			}

			// запишим увеличенное значение
			v := metrics.Counter(mdb.Delta.Int64)
			if _, err = txCounterUpdate.ExecContext(ctx, id, delta+v); err != nil {
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
