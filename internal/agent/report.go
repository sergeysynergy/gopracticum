package agent

import (
	"context"
	"log"
	"time"

	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// Выполняем регулярную отправку метрик на сервер пока не пришёл сигнал отмены.
func (a *Agent) reportTicker(ctx context.Context) {
	ticker := time.NewTicker(a.reportInterval)
	for {
		select {
		case <-ticker.C:
			a.report(ctx)
		case <-ctx.Done():
			log.Println("[INFO] Штатное завершение работы отправки метрик")
			ticker.Stop()
			return
		}
	}
}

// Выполняем отправку запросов метрик на сервер.
func (a *Agent) report(ctx context.Context) {
	hm := make([]metrics.Metrics, 0, metrics.TypeGaugeLen+metrics.TypeCounterLen)

	prm, err := a.storage.GetMetrics()
	if err != nil {
		a.handleError(err)
		return
	}

	var hash string

	for k, v := range prm.Gauges {
		value := float64(v)

		// добавляем хэш, если задан ключ key
		if a.key != "" {
			hash = metrics.GaugeHash(a.key, k, value)
		}

		hm = append(hm, metrics.Metrics{
			ID:    k,
			MType: metrics.TypeGauge,
			Value: &value,
			Hash:  hash,
		})
	}

	for k, v := range prm.Counters {
		delta := int64(v)

		// добавляем хэш, если задан ключ key
		if a.key != "" {
			hash = metrics.CounterHash(a.key, k, delta)
		}

		hm = append(hm, metrics.Metrics{
			ID:    k,
			MType: metrics.TypeCounter,
			Delta: &delta,
			Hash:  hash,
		})
	}

	if len(hm) == 0 {
		log.Println("[WARNING] Пустой массив метрик, отправлять нечего")
		return
	}

	if a.grpc {
		a.sendGRPCReport(hm)
	} else {
		_, err = a.sendHTTPReport(ctx, hm)
	}
	if err != nil {
		a.handleError(err)
		return
	}

	log.Println("[INFO] Выполнена отправка отчёта")
}
