package handlers

import (
	"encoding/json"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
	"io/ioutil"
	"net/http"
)

func (h *Handler) Value(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != "application/json" {
		h.errorJsonUnsupportedMediaType(w)
		return
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.errorJsonReadBodyFailed(w, err)
		return
	}
	defer r.Body.Close()

	m := metrics.Metrics{}
	err = json.Unmarshal(reqBody, &m)
	if err != nil {
		h.errorJsonUnmarshalFailed(w, err)
		return
	}

	switch m.MType {
	case "":
		h.errorJson(w, "Metric type needed", http.StatusBadRequest)
		return
	case "gauge":
		gauge, errGet := h.storage.GetGauge(m.ID)
		if errGet != nil {
			h.errorJson(w, errGet.Error(), http.StatusNotFound)
			return
		}
		val := float64(gauge)
		m.Value = &val
	case "counter":
		counter, errGet := h.storage.GetCounter(m.ID)
		if errGet != nil {
			h.errorJson(w, errGet.Error(), http.StatusNotFound)
			return
		}
		delta := int64(counter)
		m.Delta = &delta
	default:
		h.errorJson(w, "Given metric type not implemented", http.StatusNotImplemented)
		return
	}

	body, err := json.Marshal(&m)
	if err != nil {
		h.errorJsonMarshalFailed(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
