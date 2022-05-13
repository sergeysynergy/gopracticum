package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

func (h *Handler) Updates(w http.ResponseWriter, r *http.Request) {
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

	mcs := make([]metrics.Metrics, 0)
	err = json.Unmarshal(reqBody, &mcs)
	if err != nil {
		h.errorJSONUnmarshalFailed(w, err)
		return
	}

	prm := metrics.NewProxyMetrics()
	for _, m := range mcs {
		switch m.MType {
		case "gauge":
			if m.Value == nil {
				h.errorJSON(w, "nil gauge value", http.StatusBadRequest)
				return
			}

			if h.key != "" && m.Hash != "" {
				if metrics.GaugeHash(h.key, m.ID, *m.Value) != m.Hash {
					err = fmt.Errorf("hash check failed for gauge metric")
					h.errorJSON(w, err.Error(), http.StatusBadRequest)
					return
				}
			}

			prm.Gauges[m.ID] = metrics.Gauge(*m.Value)
		case "counter":
			if m.Delta == nil {
				h.errorJSON(w, "nil counter value", http.StatusBadRequest)
				return
			}

			if h.key != "" && m.Hash != "" {
				if metrics.CounterHash(h.key, m.ID, *m.Delta) != m.Hash {
					err = fmt.Errorf("hash check failed for counter metric")
					h.errorJSON(w, err.Error(), http.StatusBadRequest)
					return
				}
			}

			prm.Counters[m.ID] = metrics.Counter(*m.Delta)
		default:
			err = fmt.Errorf("not implemented")
			h.errorJSON(w, err.Error(), http.StatusNotImplemented)
			return
		}
	}

	err = h.storer.PutMetrics(prm)
	if err != nil {
		h.errorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
