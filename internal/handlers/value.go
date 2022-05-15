package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

func (h *Handler) Value(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != applicationJSON {
		h.errorJSONUnsupportedMediaType(w)
		return
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.errorJSONReadBodyFailed(w, err)
		return
	}
	defer r.Body.Close()

	m := metrics.Metrics{}
	err = json.Unmarshal(reqBody, &m)
	if err != nil {
		h.errorJSONUnmarshalFailed(w, err)
		return
	}

	switch m.MType {
	case "":
		h.errorJSON(w, "Metric type needed", http.StatusBadRequest)
		return
	case "gauge":
		gauge, errGet := h.storer.Get(m.ID)
		if errGet != nil {
			msg := fmt.Sprintf("%s; type: gauge; id: %s", errGet, m.ID)
			h.errorJSON(w, msg, http.StatusNotFound)
			return
		}
		value := float64(gauge.(metrics.Gauge))
		m.Value = &value

		// добавим хэш в ответ при наличии ключа
		if h.key != "" {
			m.Hash = metrics.GaugeHash(h.key, m.ID, *m.Value)
		}
	case "counter":
		counter, errGet := h.storer.Get(m.ID)
		if errGet != nil {
			msg := fmt.Sprintf("%s; type: counter; id: %s", errGet, m.ID)
			h.errorJSON(w, msg, http.StatusNotFound)
			return
		}
		delta := int64(counter.(metrics.Counter))
		m.Delta = &delta

		// добавим хэш в ответ при наличии ключа
		if h.key != "" {
			m.Hash = metrics.CounterHash(h.key, m.ID, *m.Delta)
		}
	default:
		h.errorJSON(w, "Given metric type not implemented", http.StatusNotImplemented)
		return
	}

	body, err := json.Marshal(&m)
	if err != nil {
		h.errorJSONMarshalFailed(w, err)
		return
	}

	w.Header().Set("Content-Type", applicationJSON)
	w.Write(body)

	// добавим вывод запрошенных значений в лог
	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		switch m.MType {
		case "gauge":
			log.Printf("[%s] \"POST http://%s/value/\" type - gauge; id - %s; value - %d", reqID, r.Host, m.ID, m.Value)
		case "counter":
			log.Printf("[%s] \"POST http://%s/value/\" type - counter; id - %s; delta - %d", reqID, r.Host, m.ID, m.Delta)
		default:
		}
	}
}
