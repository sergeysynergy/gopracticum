package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"io/ioutil"
	"log"
	"net/http"
)

// Updates Массово добавляет или обновляет значение метрик в хранилище.
func (h *Handler) Updates(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("\"POST http://%s/value/\"", r.Host)
	prefix := fmt.Sprintf("[%s]", middleware.GetReqID(r.Context()))
	log.Printf("%s [DEBUG] %s bulk updates metrics", prefix, url)

	ct := r.Header.Get("Content-Type")
	if ct != applicationJSON {
		h.errorJSONUnsupportedMediaType(w, r)
		return
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.errorJSONReadBodyFailed(w, r, err)
		return
	}
	defer r.Body.Close()

	mcs := make([]metrics.Metrics, 0)
	err = json.Unmarshal(reqBody, &mcs)
	if err != nil {
		h.errorJSONUnmarshalFailed(w, r, err)
		return
	}
	log.Printf("%s [DEBUG] %s total metrics to update %d", prefix, url, len(mcs))

	prm := metrics.NewProxyMetrics()
	for _, m := range mcs {
		switch m.MType {
		case "gauge":
			if m.Value == nil {
				h.errorJSON(w, r, "nil gauge value", http.StatusBadRequest)
				return
			}

			if h.key != "" && m.Hash != "" {
				if metrics.GaugeHash(h.key, m.ID, *m.Value) != m.Hash {
					err = fmt.Errorf("hash check failed for gauge metric")
					h.errorJSON(w, r, err.Error(), http.StatusBadRequest)
					return
				}
			}

			prm.Gauges[m.ID] = metrics.Gauge(*m.Value)
		case "counter":
			if m.Delta == nil {
				h.errorJSON(w, r, "nil counter value", http.StatusBadRequest)
				return
			}

			if h.key != "" && m.Hash != "" {
				if metrics.CounterHash(h.key, m.ID, *m.Delta) != m.Hash {
					err = fmt.Errorf("hash check failed for counter metric")
					h.errorJSON(w, r, err.Error(), http.StatusBadRequest)
					return
				}
			}

			// проверяем и суммируем дублирующие значения
			v, ok := prm.Counters[m.ID]
			if !ok {
				prm.Counters[m.ID] = metrics.Counter(*m.Delta)
			} else {
				prm.Counters[m.ID] = v + metrics.Counter(*m.Delta)
			}
		default:
			err = fmt.Errorf("not implemented")
			h.errorJSON(w, r, err.Error(), http.StatusNotImplemented)
			return
		}
	}

	err = h.uc.PutMetrics(prm)
	if err != nil {
		h.errorJSON(w, r, err.Error(), http.StatusBadRequest)
		return
	}
}
