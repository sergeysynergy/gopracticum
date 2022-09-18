// Package handlers Пакет реализует JSON API и точки подключения `endpoints` для работы с хранилищем метрик.
package handlers

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sergeysynergy/metricser/internal/storage"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"
)

const (
	applicationJSON = "application/json"
	textHTML        = "text/html"
)

// Handler хранит объекты роутов, репозитория и непосредственно бизнес-логики для работы с хранилищем метрик.
type Handler struct {
	router        chi.Router
	storer        storage.Repo
	fileStorer    storage.FileStorer
	dbStorer      storage.DBStorer
	key           string
	privateKey    *rsa.PrivateKey
	trustedSubnet *net.IPNet
}

type Option func(handler *Handler)

// New Создаёт новый объект JSON API Handler.
func New(opts ...Option) *Handler {
	h := &Handler{
		router:        chi.NewRouter(), // создадим новый роутер
		trustedSubnet: &net.IPNet{},    // создадим объект доверенной сети
	}

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

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	h.router.Use(cidrCheck(h.trustedSubnet))
	h.router.Use(gzipDecompressor)
	h.router.Use(gzipCompressor)
	h.router.Use(decrypt(h.privateKey))
	h.router.Use(middleware.RequestID)
	h.router.Use(middleware.RealIP)
	h.router.Use(middleware.Logger)
	h.router.Use(middleware.Recoverer)

	// определим маршруты
	h.setRoutes()

	// вернуть измененный экземпляр Handler
	return h
}

func WithTrustedSubnet(cidr string) Option {
	return func(h *Handler) {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Println("[ERROR] Failed to parse CIDR! Trusted network will NOT BE used -", err)
			return
		}
		h.trustedSubnet = ipNet
	}
}

func WithPrivateKey(key *rsa.PrivateKey) Option {
	return func(h *Handler) {
		h.privateKey = key
	}
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
