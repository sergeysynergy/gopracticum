package pgsql

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
)

func (r *Repo) Restore(m metrics.ProxyMetrics) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txGaugeUpdate := tx.StmtContext(r.ctx, r.stmtGaugeUpdate)
	txCounterUpdate := tx.StmtContext(r.ctx, r.stmtCounterUpdate)
	txGaugeInsert := tx.StmtContext(r.ctx, r.stmtGaugeInsert)
	txCounterInsert := tx.StmtContext(r.ctx, r.stmtCounterInsert)

	// запишем значения gauge
	if m.Gauges != nil {
		for id, value := range m.Gauges {
			var errGauge error
			result, errGauge := txGaugeUpdate.Exec(id, value)
			if errGauge != nil {
				return err
			}
			count, errGauge := result.RowsAffected()
			if errGauge != nil {
				return err
			}
			if count == 0 {
				_, errGaugeInsert := txGaugeInsert.ExecContext(r.ctx, id, value)
				if errGaugeInsert != nil {
					return err
				}
			}
		}
	}

	// запишем значения counters
	if m.Counters != nil {
		for id, delta := range m.Counters {
			var errCounter error
			result, errCounter := txCounterUpdate.Exec(id, delta)
			if errCounter != nil {
				return err
			}
			count, errCounter := result.RowsAffected()
			if errCounter != nil {
				return err
			}
			if count == 0 {
				_, errCounterInsert := txCounterInsert.ExecContext(r.ctx, id, delta)
				if errCounterInsert != nil {
					return err
				}
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
