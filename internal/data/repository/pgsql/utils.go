package pgsql

import (
	"database/sql"
	"github.com/sergeysynergy/metricser/internal/data/model"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// closeStatements Закрывает SLQ-утверждения при завершении работы.
func (r *Repo) closeStatements() error {
	err := r.stmtGaugeInsert.Close()
	if err != nil {
		return err
	}

	err = r.stmtCounterInsert.Close()
	if err != nil {
		return err
	}

	err = r.stmtGaugeUpdate.Close()
	if err != nil {
		return err
	}

	err = r.stmtCounterUpdate.Close()
	if err != nil {
		return err
	}

	err = r.stmtGaugeGet.Close()
	if err != nil {
		return err
	}

	err = r.stmtCounterGet.Close()
	if err != nil {
		return err
	}

	err = r.stmtAllUpdate.Close()
	if err != nil {
		return err
	}

	err = r.stmtAllSelect.Close()
	if err != nil {
		return err
	}

	return nil
}

// checkCounter Получает текущее значение счётчика.
func (r *Repo) checkCounter(id string) (metrics.Counter, error) {
	m := model.Metrics{}
	row := r.stmtCounterGet.QueryRowContext(r.ctx, id)

	err := row.Scan(&m.ID, &m.MType, &m.Value, &m.Delta)
	if err == sql.ErrNoRows {
		_, err = r.stmtCounterInsert.ExecContext(r.ctx, id, 0)
		if err != nil {
			return 0, err
		}
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return metrics.Counter(m.Delta.Int64), nil
}
