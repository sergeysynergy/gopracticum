package httpserver

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sergeysynergy/gopracticum/internal/storage"
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

func New(cfg Config) *Server {
	// определяем общее хранилище метрик
	st := storage.New()

	// задаём обработчики с доступом к общему хранилищу
	handler := &Handler{Storage: st}

	// созданим новый роутер
	r := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// определим маршруты
	getRoutes(r, handler)

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

// объявим роуты, используя маршрутизатор chi
func getRoutes(r chi.Router, handler *Handler) chi.Router {
	r.Get("/", handler.List)

	// шаблон роутов POST http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	r.Post("/update/{type}/{name}/{value}", handler.Post)

	// шаблон роутов GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>
	r.Get("/value/{type}/{name}", handler.Get)

	return r
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
