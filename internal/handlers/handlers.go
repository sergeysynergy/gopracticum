package handlers

import (
	"bytes"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
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
	router chi.Router
	storer storage.RepoStorer
	key    string
}

type Options func(handler *Handler)

func New(opts ...Options) *Handler {
	h := &Handler{
		// созданим новый роутер
		router: chi.NewRouter(),
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

	// создадим хранилище метрик реализующее интерфейс RepoStorer, если оно не было проинициализировано через WithRepoStorer
	if h.storer == nil {
		h.storer = filestore.New()
	}

	// вернуть измененный экземпляр Handler
	return h
}

func WithRepoStorer(st storage.RepoStorer) Options {
	return func(h *Handler) {
		h.storer = st
	}
}

func WithKey(key string) Options {
	return func(h *Handler) {
		h.key = key
	}
}

func (h *Handler) GetRouter() chi.Router {
	return h.router
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
	for k, val := range h.storer.GetMetrics().Gauges {
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
	for k, val := range h.storer.GetMetrics().Counters {
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
		mac2 := metrics.GaugeHash(h.key, m.ID, *m.Value)
		if mac1 != mac2 {
			log.Printf(":: mac1 - %s\n", mac1)
			log.Printf(":: mac2 - %s\n", mac2)
			return fmt.Errorf("hash check failed for gauge metric")
		}
	case "counter":
		if m.Delta == nil {
			m.Delta = new(int64)
		}

		mac1 := m.Hash
		mac2 := metrics.CounterHash(h.key, m.ID, *m.Delta)
		if mac1 != mac2 {
			return fmt.Errorf("hash check failed for counter metric")
		}
	default:
		err := fmt.Errorf("not implemented")
		return err
	}
	return nil
}
