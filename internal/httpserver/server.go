package httpserver

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sergeysynergy/gopracticum/internal/filestore"
)

type Server struct {
	server       *http.Server
	fileStore    *filestore.FileStore
	graceTimeout time.Duration // время на штатное завершения работы сервера
}

type Option func(server *Server)

func New(h http.Handler, fs *filestore.FileStore, opts ...Option) *Server {
	const (
		defaultAddress      = "127.0.0.1:8080"
		defaultGraceTimeout = 20 * time.Second
	)
	s := &Server{
		server: &http.Server{
			Addr:           defaultAddress,
			ReadTimeout:    time.Second * 10,
			WriteTimeout:   time.Second * 10,
			IdleTimeout:    time.Second * 10, // максимальное время ожидания следующего запроса
			MaxHeaderBytes: 1 << 20,          // 2^20 = 128 Kb
			Handler:        h,
		},
		fileStore:    fs,
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

func WithAddress(addr string) Option {
	return func(s *Server) {
		if addr != "" {
			s.server.Addr = addr
		}
	}
}

func (s *Server) Serve() {
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

		// штатно завершим работу рутины периодического сохранения данных метрик в файл
		err := s.fileStore.Shutdown()
		if err != nil {
			log.Fatal("[ERROR] File store shutdown error - ", err)
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

	// вызовем рутину периодического сохранения данных метрик в файл
	if s.fileStore != nil {
		s.fileStore.WriteTicker()
	}

	// запустим сервер
	log.Printf("starting HTTP-server at %s\n", s.server.Addr)
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("[ERROR] Failed to run HTTP-server", err)
	}
}
