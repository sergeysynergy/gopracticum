package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
	"io/ioutil"
	"net/http"
)

func (h *Handler) Value(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ct := r.Header.Get("Content-Type")
	if ct != "application/json" {
		msg := "wrong content type, 'application/json' needed"
		http.Error(w, msg, http.StatusUnsupportedMediaType)
		return
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer r.Body.Close()

	m := metrics.Metrics{}
	err = json.Unmarshal(reqBody, &m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch m.MType {
	case "gauge":
		gauge, errGet := h.storage.GetGauge(m.ID)
		if errGet != nil {
			http.Error(w, errGet.Error(), http.StatusNotFound)
			return
		}
		val := float64(gauge)
		m.Value = &val
	case "counter":
		h.storage.PostCounter(m.ID, metrics.Counter(*m.Delta))
	default:
		err = fmt.Errorf("not implemented")
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}

	body, err := json.Marshal(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
