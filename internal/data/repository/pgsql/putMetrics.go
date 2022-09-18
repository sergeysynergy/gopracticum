package pgsql

import (
	"database/sql"
	"github.com/sergeysynergy/metricser/internal/data/model"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
)

// PutMetrics Массово записывает значение метрик в БД.
func (r *Repo) PutMetrics(prm *metrics.ProxyMetrics) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txGaugeUpdate := tx.StmtContext(r.ctx, r.stmtGaugeUpdate)
	txCounterUpdate := tx.StmtContext(r.ctx, r.stmtCounterUpdate)
	txGaugeInsert := tx.StmtContext(r.ctx, r.stmtGaugeInsert)
	txCounterInsert := tx.StmtContext(r.ctx, r.stmtCounterInsert)
	txCounterGet := tx.StmtContext(r.ctx, r.stmtCounterGet)

	if prm.Gauges != nil {
		for id, value := range prm.Gauges {
			result, errGauge := txGaugeUpdate.Exec(id, value)
			if errGauge != nil {
				return errGauge
			}
			count, errGauge := result.RowsAffected()
			if errGauge != nil {
				return errGauge
			}
			if count == 0 {
				_, errGaugeInsert := txGaugeInsert.ExecContext(r.ctx, id, value)
				if errGaugeInsert != nil {
					return errGaugeInsert
				}
			}
		}
	}

	if prm.Counters != nil {
		for id, delta := range prm.Counters {
			// получим текущее значение счётчика
			mtx := model.Metrics{}
			// s.pgsql.PrepareContext(s.ctx, "SELECT id, type, value, delta FROM metrics WHERE id=$1")
			row := txCounterGet.QueryRowContext(r.ctx, id)
			err = row.Scan(&mtx.ID, &mtx.MType, &mtx.Value, &mtx.Delta)
			if err == sql.ErrNoRows {
				// добавим новую запись в случае отсутствия результата
				// s.pgsql.PrepareContext(s.ctx, "INSERT INTO metrics (id, type, delta) VALUES ($1, 'counter', $2)")
				_, err = txCounterInsert.ExecContext(r.ctx, id, delta)
				if err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}

			// запишем увеличенное значение
			v := metrics.Counter(0)
			if mtx.Delta.Valid {
				v = metrics.Counter(mtx.Delta.Int64)
			}
			hm := delta + v
			// s.pgsql.PrepareContext(s.ctx, "UPDATE metrics SET delta = $2 WHERE id = $1")
			if _, err = txCounterUpdate.ExecContext(r.ctx, id, hm); err != nil {
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
