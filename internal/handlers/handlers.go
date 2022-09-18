// Package handlers Пакет реализует JSON API и точки подключения `endpoints` для работы с хранилищем метрик.
package handlers

import (
	"crypto/rsa"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sergeysynergy/metricser/internal/domain/storage"
	"log"
	"net"
)

const (
	applicationJSON = "application/json"
	textHTML        = "text/html"
)

// Handler хранит объекты роутов, репозитория и непосредственно бизнес-логики для работы с хранилищем метрик.
type Handler struct {
	router chi.Router

	//storer        storage2.Repo
	//fileStorer    storage2.FileStorer
	//dbStorer      storage2.DBStorer
	uc storage.UseCase

	key           string
	privateKey    *rsa.PrivateKey
	trustedSubnet *net.IPNet
}

type Option func(handler *Handler)

// New Создаёт новый объект JSON API Handler.
func New(uc storage.UseCase, opts ...Option) *Handler {
	h := &Handler{
		router:        chi.NewRouter(), // создадим новый роутер
		trustedSubnet: &net.IPNet{},    // создадим объект доверенной сети
		uc:            uc,
	}

	// применяем в цикле каждую опцию
	for _, opt := range opts {
		// *Handler как аргумент
		opt(h)
	}

	// проинициализируем хранилище Storer
	//if h.dbStorer != nil {
	//	h.storer = h.dbStorer
	//	log.Println("database storer chosen")
	//} else if h.fileStorer != nil {
	//	log.Println("filestore storer chosen")
	//	h.storer = h.fileStorer
	//} else {
	//	log.Println("default storer chosen")
	//	h.storer = storage2.New()
	//}

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
//func WithFileStorer(fs storage2.FileStorer) Option {
//	return func(h *Handler) {
//		if fs != nil {
//			log.Println("file storage plugin connected")
//			h.fileStorer = fs
//		}
//	}
//}

// WithDBStorer Использует переданный репозиторий.
//func WithDBStorer(db storage2.DBStorer) Option {
//	return func(h *Handler) {
//		if db != nil {
//			log.Println("database plugin connected")
//			h.dbStorer = db
//		}
//	}
//}

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
