package httpserver

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	Address      string
	Port         string
	GraceTimeout time.Duration // время на штатное завершения работы сервера
}

type Server struct {
	*http.Server
	Cfg Config
}

func New(r http.Handler, cfg Config) *Server {
	// объявим HTTP-сервер
	addr := cfg.Address + ":" + cfg.Port
	s := &Server{
		Cfg: cfg,
		Server: &http.Server{
			Addr:           addr,
			ReadTimeout:    time.Second * 10,
			WriteTimeout:   time.Second * 10,
			IdleTimeout:    time.Second * 10, // максимальное время ожидания следующего запроса
			MaxHeaderBytes: 1 << 20,          // 2^20 = 128 Kb
			Handler:        r,
		},
	}

	return s
}

func (s *Server) Serve() {
	// зададим контекст выполнения сервера
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// штатное завершение по сигналам: syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// определяем время для штатного завершения работы сервера
		shutdownCtx, cancel := context.WithTimeout(serverCtx, s.Cfg.GraceTimeout)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Пришёл сигнал завершить работу: штатно завершаем работу сервера не прерывая никаких активных подключений.
		// Завершение работы выполняется в порядке:
		// - закрытия всех открытых подключений;
		// - затем закрытия всех незанятых подключений;
		// - а затем бесконечного ожидания возврата подключений в режим ожидания;
		// - наконец, завершения работы.
		err := s.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// запустим сервер
	addr := s.Cfg.Address + ":" + s.Cfg.Port
	log.Printf("starting HTTP-server at %s\n", addr)
	err := s.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// ожидаем сигнала остановки сервера через context
	<-serverCtx.Done()
}
