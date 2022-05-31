package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/sergeysynergy/metricser/pkg/metrics"
)

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

		err = h.storer.Put(m.ID, metrics.Gauge(*m.Value))
		if err != nil {
			h.errorJSON(w, r, err.Error(), http.StatusBadRequest)
			return
		}
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

		err = h.storer.Put(m.ID, metrics.Counter(*m.Delta))
		if err != nil {
			h.errorJSON(w, r, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		err = fmt.Errorf("not implemented")
		h.errorJSON(w, r, err.Error(), http.StatusNotImplemented)
		return
	}

	w.WriteHeader(http.StatusOK)

	// запишем метрики в файл, если проинициализировано хранилище на базе файла
	if h.fileStorer != nil {
		_, err = h.fileStorer.WriteMetrics()
		if err != nil {
			h.errorJSON(w, r, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
