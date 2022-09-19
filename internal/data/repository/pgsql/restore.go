package pgsql

import (
	metricserErrors "github.com/sergeysynergy/metricser/internal/errors"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
)

func (s *Storage) Restore(prm *metrics.ProxyMetrics) error {
	if prm == nil {
		return metricserErrors.ErrEmptyProxyMetrics
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txGaugeUpdate := tx.StmtContext(s.ctx, s.stmtGaugeUpdate)
	txCounterUpdate := tx.StmtContext(s.ctx, s.stmtCounterUpdate)
	txGaugeInsert := tx.StmtContext(s.ctx, s.stmtGaugeInsert)
	txCounterInsert := tx.StmtContext(s.ctx, s.stmtCounterInsert)

	// запишем значения gauge
	if prm.Gauges != nil {
		for id, value := range prm.Gauges {
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
				_, errGaugeInsert := txGaugeInsert.ExecContext(s.ctx, id, value)
				if errGaugeInsert != nil {
					return err
				}
			}
		}
	}

	// запишем значения counters
	if prm.Counters != nil {
		for id, delta := range prm.Counters {
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
				_, errCounterInsert := txCounterInsert.ExecContext(s.ctx, id, delta)
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
