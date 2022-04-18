package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

func (h *Handler) Value(w http.ResponseWriter, r *http.Request) {
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

	m := metrics.Metrics{}
	err = json.Unmarshal(reqBody, &m)
	if err != nil {
		h.errorJSONUnmarshalFailed(w, err)
		return
	}

	switch m.MType {
	case "":
		h.errorJSON(w, "Metric type needed", http.StatusBadRequest)
		return
	case "gauge":
		err = h.hashCheck(&m)
		if err != nil {
			h.errorJSON(w, err.Error(), http.StatusBadRequest)
		}

		gauge, errGet := h.storage.GetGauge(m.ID)
		if errGet != nil {
			h.errorJSON(w, errGet.Error(), http.StatusNotFound)
			return
		}
		val := float64(gauge)
		m.Value = &val
	case "counter":
		err = h.hashCheck(&m)
		if err != nil {
			h.errorJSON(w, err.Error(), http.StatusBadRequest)
		}

		counter, errGet := h.storage.GetCounter(m.ID)
		if errGet != nil {
			h.errorJSON(w, errGet.Error(), http.StatusNotFound)
			return
		}
		delta := int64(counter)
		m.Delta = &delta
	default:
		h.errorJSON(w, "Given metric type not implemented", http.StatusNotImplemented)
		return
	}

	body, err := json.Marshal(&m)
	if err != nil {
		h.errorJSONMarshalFailed(w, err)
		return
	}

	w.Header().Set("Content-Type", applicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
