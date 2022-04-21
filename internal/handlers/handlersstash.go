package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	switch metricType {
	case "gauge":
		var gauge metrics.Gauge
		err := gauge.FromString(value)
		if err != nil {
			msg := fmt.Sprintf("value %v not acceptable - %v", name, err)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		h.storage.PutGauge(name, gauge)
	case "counter":
		var counter metrics.Counter
		err := counter.FromString(value)
		if err != nil {
			msg := fmt.Sprintf("value %v not acceptable - %v", name, err)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		h.storage.PostCounter(name, counter)
	default:
		err := fmt.Errorf("not implemented")
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	var val string

	switch metricType {
	case "gauge":
		gauge, err := h.storage.GetGauge(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		val = strconv.FormatFloat(float64(gauge), 'f', -1, 64)
	case "counter":
		counter, err := h.storage.GetCounter(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		val = fmt.Sprintf("%d", counter)
	default:
		err := fmt.Errorf("not implemented")
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(val))
}

