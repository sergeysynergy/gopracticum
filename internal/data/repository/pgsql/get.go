package pgsql

import (
	"fmt"
	"log"

	"github.com/sergeysynergy/metricser/internal/data/model"
	metricserErrors "github.com/sergeysynergy/metricser/internal/errors"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// Get извлекает из БД метрику любого типа по ID.
func (s *Storage) Get(id string) (interface{}, error) {
	m := model.Metrics{}
	row := s.db.QueryRowContext(
		s.ctx,
		`SELECT id, type, value, delta FROM metrics WHERE id=$1`,
		id,
	)
	// разбираем результат
	err := row.Scan(&m.ID, &m.MType, &m.Value, &m.Delta)
	if err != nil {
		return nil, err
	}

	switch m.MType {
	case metrics.TypeGauge:
		if !m.Value.Valid {
			return nil, fmt.Errorf("NULL gauge value")
		}
		return metrics.Gauge(m.Value.Float64), nil
	case metrics.TypeCounter:
		if !m.Delta.Valid {
			return nil, fmt.Errorf("NULL counter value")
		}
		return metrics.Counter(m.Delta.Int64), nil
	default:
	}

	return nil, metricserErrors.MetricNotImplemented
}

// GetMetrics извлекает из БД значение всех метрик.
func (s *Storage) GetMetrics() (metrics.ProxyMetrics, error) {
	mcs := metrics.NewProxyMetrics()

	//ctx, cancel := context.WithTimeout(parentCtx, queryTimeOut)
	//defer cancel()

	rows, err := s.db.QueryContext(s.ctx, `SELECT id, type, value, delta FROM metrics`)
	if err != nil {
		return mcs, err
	}
	// обязательно закрываем перед возвратом функции
	defer rows.Close()

	// пробегаем по всем записям
	for rows.Next() {
		m := model.Metrics{}
		err = rows.Scan(&m.ID, &m.MType, &m.Value, &m.Delta)
		if err != nil {
			return mcs, err
		}

		switch m.MType {
		case metrics.TypeGauge:
			if !m.Value.Valid {
				log.Println("[WARNING] NULL gauge value")
			}
			mcs.Gauges[m.ID] = metrics.Gauge(m.Value.Float64)
		case metrics.TypeCounter:
			if !m.Delta.Valid {
				log.Println("[WARNING] NULL counter value")
			}
			mcs.Counters[m.ID] = metrics.Counter(m.Delta.Int64)
		default:
			log.Println("[WARNING] not implemented metrics type")
		}
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return mcs, err
	}

	return mcs, nil
}
