package db

import (
	"context"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
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
		// запишем новое значение счётчика
		result, err := s.db.ExecContext(ctx, queryGet, id)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			_, err = s.db.ExecContext(ctx, queryInsertCounter, id, m)
			if err != nil {
				return err
			}
			return nil
		}

		// получим текущее значение счётчика
		mdb := metricsDB{}
		row := s.db.QueryRowContext(ctx, queryGet, id)
		err = row.Scan(&mdb.ID, &mdb.MType, &mdb.Value, &mdb.Delta)
		if err != nil {
			return err
		}

		// проинициализируем начальное значение счётчика
		if !mdb.Delta.Valid {
			mdb.Delta.Int64 = 0
		}

		// прибавим счётчик к текущему значению
		count := m + metrics.Counter(mdb.Delta.Int64)

		// запишем новое значение счётчика
		_, err = s.db.ExecContext(ctx, queryUpdateCounter, id, count)
		if err != nil {
			return err
		}
	default:
		return metrics.ErrNotImplemented
	}

	return nil
}

func (s *Storage) PutMetrics(ctx context.Context, m metrics.ProxyMetrics) error {
	if m.Gauges != nil {
		for id, value := range m.Gauges {
			_, err := s.db.ExecContext(ctx, queryUpdateGauge, id, value)
			if err != nil {
				return err
			}
		}
	}

	if m.Counters != nil {
		for id, delta := range m.Counters {
			_, err := s.db.ExecContext(ctx, queryUpdateCounter, id, delta)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
