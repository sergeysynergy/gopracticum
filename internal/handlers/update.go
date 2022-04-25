package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != applicationJSON {
		h.errorJSONUnsupportedMediaType(w)
		return
	}

	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.errorJSONReadBodyFailed(w, err)
		return
	}
	defer r.Body.Close()

	m := metrics.Metrics{}
	err = json.Unmarshal(respBody, &m)
	if err != nil {
		h.errorJSONUnmarshalFailed(w, err)
		return
	}

	switch m.MType {
	case "gauge":
		if h.key != "" && m.Hash != "" {
			if metrics.GaugeHash(h.key, m.ID, *m.Value) != m.Hash {
				err = fmt.Errorf("hash check failed for gauge metric")
				h.errorJSON(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		err = h.storer.Put(m.ID, metrics.Gauge(*m.Value))
		if err != nil {
			h.errorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}
	case "counter":
		if h.key != "" && m.Hash != "" {
			if metrics.CounterHash(h.key, m.ID, *m.Delta) != m.Hash {
				err = fmt.Errorf("hash check failed for counter metric")
				h.errorJSON(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		err = h.storer.Put(m.ID, metrics.Counter(*m.Delta))
		if err != nil {
			h.errorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		err = fmt.Errorf("not implemented")
		h.errorJSON(w, err.Error(), http.StatusNotImplemented)
		return
	}

	w.WriteHeader(http.StatusOK)

	// запишем метрики в файл, если проинициализировано хранилище на базе файла
	if h.fileStorer != nil {
		_, err = h.fileStorer.WriteMetrics()
		if err != nil {
			h.errorJSON(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
