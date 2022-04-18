package agent

import (
	"context"
	"fmt"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
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
