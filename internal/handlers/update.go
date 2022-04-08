package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
	"io/ioutil"
	"net/http"
)

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != "application/json" {
		msg := "wrong content type, 'application/json' needed"
		http.Error(w, msg, http.StatusUnsupportedMediaType)
		return
	}

	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	m := metrics.Metrics{}
	err = json.Unmarshal(respBody, &m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	switch m.MType {
	case "gauge":
		h.storage.PutGauge(m.ID, metrics.Gauge(*m.Value))
	case "counter":
		h.storage.PostCounter(m.ID, metrics.Counter(*m.Delta))
	default:
		err = fmt.Errorf("not implemented")
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}

	w.WriteHeader(http.StatusOK)
}
