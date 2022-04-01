package httpserver

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sergeysynergy/gopracticum/internal/handlers"
	"github.com/sergeysynergy/gopracticum/internal/storage"
)

type Config struct {
	Address string // адрес сервера куда отправлять метрики
	Port    string
}

type Server struct {
	*http.Server
	Cfg Config
	ctx context.Context
}

func New(cfg Config) *Server {
	st := storage.New()
	gaugeHandler := &handlers.Gauge{Storage: st}
	counterHandler := &handlers.Counter{Storage: st}
	checkHandler := &handlers.Check{Storage: st}

	// Шаблон роутов http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	// Объявляем роуты используя штатный маршрутизатор пакета http.
	mux := http.NewServeMux()
	mux.Handle("/update/gauge/", gaugeHandler)
	mux.Handle("/update/counter/", counterHandler)
	mux.Handle("/check", checkHandler)
	mux.HandleFunc("/update/", handlers.NotImplemented)

	// Добавляем middleware.
	mainHandler := http.Handler(mux)
	mainHandler = preChecksMiddleware(mainHandler)
	mainHandler = accessLogMiddleware(mainHandler)
	mainHandler = panicMiddleware(mainHandler)

	addr := cfg.Address + ":" + cfg.Port
	s := &Server{
		Cfg: cfg,
		Server: &http.Server{
			Addr:           addr,
			ReadTimeout:    time.Second * 10,
			WriteTimeout:   time.Second * 10,
			IdleTimeout:    time.Second * 10, // максимальное время ожидания следующего запроса
			MaxHeaderBytes: 1 << 20,          // 2^20 = 128 Kb
			Handler:        mainHandler,
		},
		ctx: context.Background(),
	}

	return s
}

func (s *Server) Serve() {
	idleConnsClosed := make(chan struct{})
	go func() {
		// Штатное завершение по сигналам: syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT.
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// Пришёл сигнал завершить работу: штатно завершаем работу сервера, не прерывая никаких активных подключений.
		// Завершение работы выполняется в порядке:
		// - закрытия всех открытых подключений;
		// - затем закрытия всех незанятых подключений;
		// - а затем бесконечного ожидания возврата подключений в режим ожидания;
		// - наконец, завершения работы.
		if err := s.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("HTTP-server started at: %s\n", s.Addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}
