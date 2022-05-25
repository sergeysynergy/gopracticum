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
	url := fmt.Sprintf("\"POST http://%s/value/\"", r.Host)
	prefix := fmt.Sprintf("[%s]", middleware.GetReqID(r.Context()))

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
	log.Printf("%s [DEBUG] %s request body: %s", prefix, url, string(reqBody))

	m := metrics.Metrics{}
	err = json.Unmarshal(reqBody, &m)
	if err != nil {
		h.errorJSONUnmarshalFailed(w, r, err)
		return
	}

	switch m.MType {
	case "":
		h.errorJSON(w, r, "Metric type needed", http.StatusBadRequest)
		return
	case "gauge":
		gauge, errGet := h.storer.Get(m.ID)
		if errGet != nil {
			msg := fmt.Sprintf("%s; type: gauge; id: %s", errGet, m.ID)
			h.errorJSON(w, r, msg, http.StatusNotFound)
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
			h.errorJSON(w, r, msg, http.StatusNotFound)
			return
		}
		delta := int64(counter.(metrics.Counter))
		m.Delta = &delta

		// добавим хэш в ответ при наличии ключа
		if h.key != "" {
			m.Hash = metrics.CounterHash(h.key, m.ID, *m.Delta)
		}
	default:
		h.errorJSON(w, r, "Given metric type not implemented", http.StatusNotImplemented)
		return
	}

	body, err := json.Marshal(&m)
	if err != nil {
		h.errorJSONMarshalFailed(w, r, err)
		return
	}

	w.Header().Set("Content-Type", applicationJSON)
	w.Write(body)

	// добавим вывод запрошенных значений в лог
	switch m.MType {
	case "gauge":
		log.Printf("%s [DEBUG] %s responce body: %s", prefix, url, body)
	case "counter":
		log.Printf("%s [DEBUG] %s responce body: %s", prefix, url, body)
	default:
		log.Printf("%s [DEBUG] %s unknown metrics type", prefix, url)
	}
}
