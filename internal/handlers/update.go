package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// Update Добавляет или обновляет значение метрики в хранилище по ID.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != applicationJSON {
		h.errorJSONUnsupportedMediaType(w, r)
		return
	}

	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.errorJSONReadBodyFailed(w, r, err)
		return
	}
	defer r.Body.Close()

	m := metrics.Metrics{}
	err = json.Unmarshal(respBody, &m)
	if err != nil {
		h.errorJSONUnmarshalFailed(w, r, err)
		return
	}

	if !metrics.IsKnown(m.ID) {
		log.Println("[WARNING] unknown metric ID", m.ID)
	}

	prm := metrics.NewProxyMetrics()
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

		gauge := metrics.Gauge(*m.Value)
		err = h.uc.Put(m.ID, gauge)
		if err != nil {
			h.errorJSON(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		prm.Gauges[m.ID] = gauge
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

		counter := metrics.Counter(*m.Delta)
		err = h.uc.Put(m.ID, counter)
		if err != nil {
			h.errorJSON(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		prm.Counters[m.ID] = counter
	default:
		err = fmt.Errorf("not implemented")
		h.errorJSON(w, r, err.Error(), http.StatusNotImplemented)
		return
	}

	w.WriteHeader(http.StatusOK)

	// запишем метрики в файл
	err = h.uc.WriteMetrics(prm)
	if err != nil {
		h.errorJSON(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
}
