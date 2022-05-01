package db

import (
	"context"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
	"log"
)

func (s *Storage) Restore(ctx context.Context, m metrics.ProxyMetrics) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txGaugeUpdate := tx.StmtContext(ctx, s.stmtGaugeUpdate)
	txCounterUpdate := tx.StmtContext(ctx, s.stmtCounterUpdate)
	txGaugeInsert := tx.StmtContext(ctx, s.stmtGaugeInsert)
	txCounterInsert := tx.StmtContext(ctx, s.stmtCounterInsert)

	// запишем значения gauge
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

	// запишем значения counters
	if m.Counters != nil {
		for id, delta := range m.Counters {
			result, err := txCounterUpdate.Exec(id, delta)
			if err != nil {
				return err
			}
			count, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if count == 0 {
				_, err := txCounterInsert.ExecContext(ctx, id, delta)
				if err != nil {
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
