package db

import (
	"fmt"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
	"log"
)

func (s *Storage) Get(id string) (interface{}, error) {
	//ctx, cancel := context.WithTimeout(parentCtx, queryTimeOut)
	//defer cancel()

	m := metricsDB{}
	row := s.db.QueryRowContext(s.ctx, queryGet, id)
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

	return nil, fmt.Errorf("metric not implemented")
}

func (s *Storage) GetMetrics() (metrics.ProxyMetrics, error) {
	mcs := metrics.NewProxyMetrics()

	//ctx, cancel := context.WithTimeout(parentCtx, queryTimeOut)
	//defer cancel()

	rows, err := s.db.QueryContext(s.ctx, queryGetMetrics)
	if err != nil {
		return mcs, err
	}
	// обязательно закрываем перед возвратом функции
	defer rows.Close()

	// пробегаем по всем записям
	for rows.Next() {
		var m metricsDB
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
