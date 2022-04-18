package handlers

import (
	"bytes"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"sort"
	"strconv"

	"github.com/sergeysynergy/gopracticum/internal/filestore"
	"github.com/sergeysynergy/gopracticum/internal/storage"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

const (
	applicationJSON = "application/json"
	textHTML        = "text/html"
)

type Handler struct {
	router    chi.Router
	storage   storage.Storer
	fileStore *filestore.FileStore
	key       string
}

type Options func(handler *Handler)

func New(opts ...Options) *Handler {
	st := storage.New()
	h := &Handler{
		// созданим новый роутер
		router: chi.NewRouter(),
		// определяем хранилище метрик, реализующее интерфейс Storer
		storage: st,
	}

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	h.router.Use(gzipDecompressor)
	h.router.Use(gzipCompressor)
	h.router.Use(middleware.RequestID)
	h.router.Use(middleware.RealIP)
	h.router.Use(middleware.Logger)
	h.router.Use(middleware.Recoverer)

	// определим маршруты
	h.setRoutes()

	// применяем в цикле каждую опцию
	for _, opt := range opts {
		// *Handler как аргумент
		opt(h)
	}

	// вернуть измененный экземпляр Handler
	return h
}

func WithStorage(st storage.Storer) Options {
	return func(handler *Handler) {
		handler.storage = st
	}
}

func WithFileStore(fs *filestore.FileStore) Options {
	return func(handler *Handler) {
		handler.fileStore = fs
	}
}

func WithKey(key string) Options {
	return func(handler *Handler) {
		handler.key = key
	}
}

func (h *Handler) GetRouter() chi.Router {
	return h.router
}

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

func (h *Handler) List(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", textHTML)
	w.WriteHeader(http.StatusOK)

	var b bytes.Buffer
	b.WriteString("<h1>Current metrics data:</h1>")

	type gauge struct {
		key   string
		value float64
	}
	gauges := make([]gauge, 0, metrics.GaugeLen)
	for k, val := range h.storage.GetGauges() {
		gauges = append(gauges, gauge{key: k, value: float64(val)})
	}
	sort.Slice(gauges, func(i, j int) bool { return gauges[i].key < gauges[j].key })

	b.WriteString(`<div><h2>Gauges</h2>`)
	for _, g := range gauges {
		val := strconv.FormatFloat(g.value, 'f', -1, 64)
		b.WriteString(fmt.Sprintf("<div>%s - %v</div>", g.key, val))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<div><h2>Counters</h2>`)
	for k, val := range h.storage.GetCounters() {
		b.WriteString(fmt.Sprintf("<div>%s - %d</div>", k, val))
	}
	b.WriteString(`</div>`)

	w.Write(b.Bytes())
}

func (h *Handler) hashCheck(m *metrics.Metrics) error {
	if h.key == "" {
		return nil
	}

	switch m.MType {
	case "gauge":
		if m.Value == nil {
			m.Value = new(float64)
		}

		mac1 := m.Hash
		mac2 := metrics.GetGaugeHash(h.key, m.ID, *m.Value)
		if mac1 != mac2 {
			return fmt.Errorf("hash check failed for gauge metric")
		}
	case "counter":
		if m.Delta == nil {
			m.Delta = new(int64)
		}

		mac1 := m.Hash
		mac2 := metrics.GetCounterHash(h.key, m.ID, *m.Delta)
		if mac1 != mac2 {
			return fmt.Errorf("hash check failed for counter metric")
		}
	default:
		err := fmt.Errorf("not implemented")
		return err
	}
	return nil
}
