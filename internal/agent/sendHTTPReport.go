package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sergeysynergy/metricser/pkg/crypter"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
	"net/http"
)

// sendReport Отправляет значения всех метрик на сервер.
func (a *Agent) sendHTTPReport(ctx context.Context, hm []metrics.Metrics) (*resty.Response, error) {
	endpoint := a.protocol + a.addr + "/updates/" // адрес по которому отправляются метрики на сервер
	localIP := "127.0.0.1"                        // IP адрес клиента
	encoding := ""                                // значение указывает зашифровано ли сообщение

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
		SetHeader("X-Real-IP", localIP).
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
