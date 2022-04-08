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
