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

func (h *Handler) Updates(w http.ResponseWriter, r *http.Request) {
	prefix := fmt.Sprintf("\"POST http://%s/value/\"", r.Host)
	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		prefix = fmt.Sprintf("[%s] \"POST http://%s/value/\"", reqID, r.Host)
	}
	log.Printf("%s bulk updates metrics", prefix)

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
	log.Printf("%s total metrics to update %d", prefix, len(mcs))

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

			prm.Counters[m.ID] = metrics.Counter(*m.Delta)
		default:
			err = fmt.Errorf("not implemented")
			h.errorJSON(w, r, err.Error(), http.StatusNotImplemented)
			return
		}
	}

	err = h.storer.PutMetrics(prm)
	if err != nil {
		h.errorJSON(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
