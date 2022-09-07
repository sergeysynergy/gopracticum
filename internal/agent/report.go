package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sergeysynergy/metricser/pkg/crypter"
	"log"
	"net/http"
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

	_, err = a.sendReport(ctx, hm)
	if err != nil {
		a.handleError(err)
		return
	}

	log.Println("[INFO] Выполнена отправка отчёта")
}

func (a *Agent) sendReport(ctx context.Context, hm []metrics.Metrics) (*resty.Response, error) {
	endpoint := a.protocol + a.addr + "/updates/"
	encoding := ""
	body, err := json.Marshal(hm)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body while sending report: %w", err)
	}

	if a.publicKey != nil {
		jsonHm, errMarsh := json.Marshal(hm)
		if errMarsh != nil {
			log.Println("[WARNING] Не удалось конвертировать метрики в структуру JSON", err)
		} else {
			bodyEncrypt, errEncrypt := crypter.Encrypt(a.publicKey, jsonHm)
			if errEncrypt != nil {
				log.Println("[WARNING] Не удалось зашифровать тело запроса", err)
			} else {
				body = bodyEncrypt
				encoding = "crypted"
				log.Println("[INFO] Тело запроса было зашифровано")
			}
		}
	}

	resp, err := a.client.R().
		SetHeader("Accept", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", encoding).
		SetContext(ctx).
		SetBody(body).
		Post(endpoint)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return resp, fmt.Errorf("invalid status code %v", resp.StatusCode())
	}

	return resp, nil
}
