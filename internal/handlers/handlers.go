// Package handlers Пакет реализует JSON API и точки подключения `endpoints` для работы с хранилищем метрик.
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

	"github.com/sergeysynergy/metricser/internal/storage"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

const (
	applicationJSON = "application/json"
	textHTML        = "text/html"
)

// Handler хранит объекты роутов, репозитория и непосредственно бизнес-логики для работы с хранилищем метрик.
type Handler struct {
	router     chi.Router
	storer     storage.Repo
	fileStorer storage.FileStorer
	dbStorer   storage.DBStorer
	key        string
}

type Option func(handler *Handler)

// New Создаёт новый объект JSON API Handler.
func New(opts ...Option) *Handler {
	h := &Handler{
		// создадим новый роутер
		router: chi.NewRouter(),
	}

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	h.router.Use(gzipDecompressor)
	h.router.Use(gzipCompressor)
	h.router.Use(middleware.RequestID)
	h.router.Use(middleware.RealIP)
	h.router.Use(middleware.Logger)
	h.router.Use(middleware.Recoverer)

	// применяем в цикле каждую опцию
	for _, opt := range opts {
		// *Handler как аргумент
		opt(h)
	}

	// проинициализируем хранилище Storer
	if h.dbStorer != nil {
		h.storer = h.dbStorer
		log.Println("database storer chosen")
	} else if h.fileStorer != nil {
		log.Println("filestore storer chosen")
		h.storer = h.fileStorer
	} else {
		log.Println("default storer chosen")
		h.storer = storage.New()
	}

	// определим маршруты
	h.setRoutes()

	// вернуть измененный экземпляр Handler
	return h
}

// WithFileStorer Использует переданное файловое хранилище.
func WithFileStorer(fs storage.FileStorer) Option {
	return func(h *Handler) {
		if fs != nil {
			log.Println("file storage plugin connected")
			h.fileStorer = fs
		}
	}
}

// WithDBStorer Использует переданный репозиторий.
func WithDBStorer(db storage.DBStorer) Option {
	return func(h *Handler) {
		if db != nil {
			log.Println("database plugin connected")
			h.dbStorer = db
		}
	}
}

// WithKey Использует переданный ключ для создания хэша.
func WithKey(key string) Option {
	return func(h *Handler) {
		h.key = key
	}
}

// GetRouter Возвращает объект роутер.
func (h *Handler) GetRouter() chi.Router {
	return h.router
}

// List Возвращает список со значением всех метрик.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", textHTML)
	w.WriteHeader(http.StatusOK)

	var b bytes.Buffer
	b.WriteString("<h1>Current metrics data:</h1>")

	type gauge struct {
		key   string
		value float64
	}

	mcs, err := h.storer.GetMetrics()
	if err != nil {
		h.errorJSON(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	gauges := make([]gauge, 0, metrics.TypeGaugeLen)
	for k, val := range mcs.Gauges {
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
	for k, val := range mcs.Counters {
		b.WriteString(fmt.Sprintf("<div>%s - %d</div>", k, val))
	}
	b.WriteString(`</div>`)

	w.Write(b.Bytes())
}

// TODO: проверить зачем был нужен метод
/*
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
*/
