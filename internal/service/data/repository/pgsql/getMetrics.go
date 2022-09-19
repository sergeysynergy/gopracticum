package pgsql

import (
	"github.com/sergeysynergy/metricser/internal/service/data/model"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
)

// GetMetrics извлекает из БД значение всех метрик.
func (s *Storage) GetMetrics() (*metrics.ProxyMetrics, error) {
	prm := metrics.NewProxyMetrics()

	//ctx, cancel := context.WithTimeout(parentCtx, queryTimeOut)
	//defer cancel()

	rows, err := s.db.QueryContext(s.ctx, `SELECT id, type, value, delta FROM metrics`)
	if err != nil {
		return nil, err
	}
	// обязательно закрываем перед возвратом функции
	defer rows.Close()

	// пробегаем по всем записям
	for rows.Next() {
		m := model.Metrics{}
		err = rows.Scan(&m.ID, &m.MType, &m.Value, &m.Delta)
		if err != nil {
			return nil, err
		}

		switch m.MType {
		case metrics.TypeGauge:
			if !m.Value.Valid {
				log.Println("[WARNING] NULL gauge value")
			}
			prm.Gauges[m.ID] = metrics.Gauge(m.Value.Float64)
		case metrics.TypeCounter:
			if !m.Delta.Valid {
				log.Println("[WARNING] NULL counter value")
			}
			prm.Counters[m.ID] = metrics.Counter(m.Delta.Int64)
		default:
			log.Println("[WARNING] not implemented metrics type")
		}
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return prm, nil
}
