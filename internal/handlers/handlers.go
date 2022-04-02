package handlers

import (
	"bytes"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/sergeysynergy/gopracticum/internal/storage"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
	"net/http"
	"strconv"
)

type Handler struct {
	*storage.Storage
}

type Counter struct {
	*storage.Storage
}
type Check struct {
	*storage.Storage
}

func (h *Handler) PostGauge(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	var gauge metrics.Gauge
	err := gauge.FromString(value)
	if err != nil {
		msg := fmt.Sprintf("value %v not acceptable - %v", name, err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	h.Put(metrics.Name(name), gauge)

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetGauge(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	gauge, err := h.Storage.GetGauge(metrics.Name(name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	val := strconv.FormatFloat(float64(*gauge), 'f', -1, 64)
	w.Write([]byte(val))
}

func (h *Handler) PostCounter(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	var counter metrics.Counter
	err := counter.FromString(value)
	if err != nil {
		msg := fmt.Sprintf("value %v not acceptable - %v", name, err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	h.Count(metrics.Name(name), counter)

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetCounter(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	counter, err := h.Storage.GetCounter(metrics.Name(name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%d", *counter)))
}

func (h *Handler) List(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "text/HTML")
	w.WriteHeader(http.StatusOK)

	var b bytes.Buffer
	b.WriteString("<h1>Current metrics data:</h1>")

	b.WriteString(`<div><h2>Gauges</h2>`)
	for k, val := range h.Gauges {
		b.WriteString(fmt.Sprintf("<div>%s - %f</div>", k, val))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<div><h2>Counters</h2>`)
	for k, val := range h.Counters {
		b.WriteString(fmt.Sprintf("<div>%s - %d</div>", k, val))
	}
	b.WriteString(`</div>`)

	w.Write(b.Bytes())
}

func NotImplemented(w http.ResponseWriter, _ *http.Request) {
	err := fmt.Errorf("not implemented")
	http.Error(w, err.Error(), http.StatusNotImplemented)
}
