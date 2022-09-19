// Package httpserver Пакет предназначен для запуска сервиса по сбору и хранения метрик на базе http-сервера.
package httpserver

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/sergeysynergy/metricser/internal/storage"
)

// Server Хранит данные и объекты для реализации работы http-сервера.
type Server struct {
	server *http.Server
	uc     storage.UseCase
}

type Option func(server *Server)

// New Создаёт новый объект типа Server.
func New(uc storage.UseCase, h http.Handler, opts ...Option) *Server {
	const (
		defaultAddress = "127.0.0.1:8080"
	)
	s := &Server{
		uc: uc,
		server: &http.Server{
			Addr:           defaultAddress,
			ReadTimeout:    time.Second * 10,
			WriteTimeout:   time.Second * 10,
			IdleTimeout:    time.Second * 10, // максимальное время ожидания следующего запроса
			MaxHeaderBytes: 1 << 20,          // 2^20 = 128 Kb
			Handler:        h,
		},
	}
	// Применяем в цикле каждую опцию
	for _, opt := range opts {
		// вызываем функцию, предоставляющую экземпляр *Server в качестве аргумента
		opt(s)
	}

	// вернуть измененный экземпляр Server
	return s
}

// WithAddress Использует переданный адрес для запуска сервера.
func WithAddress(addr string) Option {
	return func(s *Server) {
		if addr != "" {
			s.server.Addr = addr
		}
	}
}

// Serve Запускает основные методы всего сервиса `Metricser`.
func (s *Server) Serve() {
	// вызовем рутину периодического сохранения данных метрик в файл, если хранилище проинициализировано
	err := s.uc.WriteTicker(nil)
	if err != nil {
		log.Println("[WARNING] Failed to start repository writing ticker - ", err)
	}

	// запустим сервер
	log.Printf("starting HTTP-server at %s\n", s.server.Addr)
	err = s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("[ERROR] Failed to run HTTP-server", err)
	}
}

func (s *Server) Shutdown() error {
	err := s.server.Shutdown(context.Background())
	if err != nil {
		return err
	}

	log.Println("[DEBUG] Gracefully shutdown http-server")
	return nil
}
