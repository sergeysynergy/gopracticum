package agent

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
	"net/http"
	"sync"
)

// старый вид отправки запроса - обычным текстом
func (a *Agent) sendBasicRequest(ctx context.Context, wg *sync.WaitGroup, key string, value interface{}) {
	defer wg.Done()

	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	var endpoint string

	switch metric := value.(type) {
	case metrics.Gauge:
		endpoint = fmt.Sprintf("%s%s/update/%s/%s/%f", a.protocol, a.addr, "gauge", key, metric)
	case metrics.Counter:
		endpoint = fmt.Sprintf("%s%s/update/%s/%s/%d", a.protocol, a.addr, "counter", key, metric)
	default:
		a.handleError(fmt.Errorf("неизвестный тип метрики"))
		return
	}

	resp, err := a.client.R().
		SetContext(ctx).
		Post(endpoint)

	if err != nil {
		a.handleError(err)
		return
	}

	if resp.StatusCode() != http.StatusOK {
		respErr := fmt.Errorf("%v", resp.StatusCode())
		a.handleError(respErr)
		return
	}
}

func (a *Agent) sendUpdate(ctx context.Context, m *metrics.Metrics) (*resty.Response, error) {
	endpoint := a.protocol + a.addr + "/update/"

	resp, err := a.client.R().
		SetHeader("Accept", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetContext(ctx).
		SetBody(m).
		Post(endpoint)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return resp, fmt.Errorf("invalid status code %v", resp.StatusCode())
	}

	return resp, nil
}

func (a *Agent) sendReportUpdate(ctx context.Context) {
	mcs, err := a.storage.GetMetrics()
	if err != nil {
		a.handleError(err)
		return
	}

	for k, v := range mcs.Gauges {
		gauge := float64(v)
		m := &metrics.Metrics{
			ID:    k,
			MType: "gauge",
			Value: &gauge,
		}

		// добавляем хэш, если задан ключ key
		if a.key != "" {
			m.Hash = metrics.GaugeHash(a.key, m.ID, *m.Value)
		}

		_, err := a.sendUpdate(ctx, m)
		if err != nil {
			a.handleError(err)
			return
		}
	}

	for k, v := range mcs.Counters {
		counter := int64(v)
		m := &metrics.Metrics{
			ID:    k,
			MType: "counter",
			Delta: &counter,
		}

		// добавляем хэш, если задан ключ key
		if a.key != "" {
			m.Hash = metrics.CounterHash(a.key, m.ID, *m.Delta)
		}

		_, err := a.sendUpdate(ctx, m)
		if err != nil {
			a.handleError(err)
			return
		}
	}

	log.Println("Выполнена отправка отчёта")
}
