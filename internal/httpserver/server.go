package httpserver

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sergeysynergy/gopracticum/internal/handlers"
	"github.com/sergeysynergy/gopracticum/internal/storage"
)

type Config struct {
	Address         string // адрес сервера куда отправлять метрики
	Port            string
	ShutdownTimeout time.Duration // отсрочка завершения работы по сигналу прерывания
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
	go func() {
		log.Printf("HTTP-server started at: %s\n", s.Addr)
		log.Fatal(s.ListenAndServe())
	}()

	// Штатное завершение по сигналам: syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-stop
	log.Println("Got server shutdown signal:", sig)
	log.Printf("Awaiting %v seconds for shutdown", s.Cfg.ShutdownTimeout)
	time.Sleep(s.Cfg.ShutdownTimeout)

	err := s.Shutdown(s.ctx)
	if err != nil {
		log.Println("[ERROR]", err.Error())
		return
	}
}
