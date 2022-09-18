// Package httpserver Пакет предназначен для запуска сервиса по сбору и хранения метрик на базе http-сервера.
package httpserver

import (
	"context"
	"github.com/sergeysynergy/metricser/internal/domain/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server Хранит данные и объекты для реализации работы http-сервера.
type Server struct {
	uc     storage.UseCase
	server *http.Server
	//fileStorer   storage.FileStorer
	//dbStorer     storage.DBStorer
	graceTimeout time.Duration // Время на штатное завершения работы сервера
}

type Option func(server *Server)

// New Создаёт новый объект типа Server.
func New(uc storage.UseCase, h http.Handler, opts ...Option) *Server {
	const (
		defaultAddress      = "127.0.0.1:8080"
		defaultGraceTimeout = 20 * time.Second
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
		graceTimeout: defaultGraceTimeout,
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

// WithFileStorer Использует переданное файловое хранилище.
//func WithFileStorer(fs storage.FileStorer) Option {
//	return func(s *Server) {
//		if fs != nil {
//			s.fileStorer = fs
//		}
//	}
//}

// WithDBStorer Использует переданный репозиторий.
//func WithDBStorer(ds storage.DBStorer) Option {
//	return func(s *Server) {
//		if ds != nil {
//			s.dbStorer = ds
//		}
//	}
//}

// graceDown Штатное завершение работы сервера.
func (s *Server) graceDown() {
	// Рутина для штатного завершения работы
	go func() {
		// штатное завершение по сигналам: syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-sig

		// определяем время для штатного завершения работы сервера
		// необходимо на случай вечного ожидания закрытия всех подключений к серверу
		shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), s.graceTimeout)
		defer shutdownCtxCancel()
		// принудительно завершаем работу по истечении срока s.graceTimeout
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("[ERROR] Graceful shutdown timed out! Forcing exit.")
			}
		}()

		// штатно завершим работу файлового хранилища, если оно проинициализировано
		//if s.fileStorer != nil {
		//	err := s.fileStorer.Shutdown()
		//	if err != nil {
		//		log.Fatal("[ERROR] Repository shutdown error - ", err)
		//	}
		//}

		// штатно завершим работу с базой данных, если она подключена
		//if s.dbStorer != nil {
		//	err := s.dbStorer.Shutdown()
		//	if err != nil {
		//		log.Fatal("[ERROR] Database shutdown error - ", err)
		//	}
		//}

		err := s.uc.Shutdown()
		if err != nil {
			log.Fatal("[ERROR] Shutdown failed - ", err)
		}

		// Штатно завершаем работу HTTP-сервера не прерывая никаких активных подключений.
		// Завершение работы выполняется в порядке:
		// - закрытия всех открытых подключений;
		// - затем закрытия всех незанятых подключений;
		// - а затем бесконечного ожидания возврата подключений в режим ожидания;
		// - наконец, завершения работы.
		err = s.server.Shutdown(context.Background())
		if err != nil {
			log.Fatal("[ERROR] Server shutdown error - ", err)
		}
	}()
}

// Serve Запускает основные методы всего сервиса `Metricser`.
func (s *Server) Serve() {
	s.graceDown()

	// вызовем рутину периодического сохранения данных метрик в файл, если хранилище проинициализировано
	//if s.fileStorer != nil {
	//	err := s.fileStorer.WriteTicker()
	//	if err != nil {
	//		log.Println("[WARNING] Failed to start repository writing ticker - ", err)
	//	}
	//}
	errLog := s.uc.WriteTicker()
	if errLog != nil {
		log.Println("[WARNING] Failed to start file writing ticker - ", errLog)
	}

	// запустим сервер
	log.Printf("starting HTTP-server at %s\n", s.server.Addr)
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("[ERROR] Failed to run HTTP-server", err)
	}
}
