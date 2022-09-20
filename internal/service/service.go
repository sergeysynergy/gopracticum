package service

import (
	"context"
	"crypto/rsa"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sergeysynergy/metricser/config"
	serviceConst "github.com/sergeysynergy/metricser/internal/service/consts"
	serviceHTTP "github.com/sergeysynergy/metricser/internal/service/delivery/http"
	"github.com/sergeysynergy/metricser/internal/service/delivery/http/handlers"
	"github.com/sergeysynergy/metricser/internal/service/storage"
)

type Service struct {
	cfg        *config.ServerConf
	privateKey *rsa.PrivateKey
	uc         storage.UseCase
	httpServer *serviceHTTP.Server
}

func New(cfg *config.ServerConf, uc storage.UseCase) *Service {
	m := &Service{
		cfg: cfg,
		uc:  uc,
	}

	m.init()

	return m
}

func (s *Service) init() {
	s.initHTTPServer()
}

// initHTTPServer Проинициализируем http-сервер.
func (s *Service) initHTTPServer() {
	// Получим обработчики для http-сервера
	h := handlers.New(s.uc,
		handlers.WithKey(s.cfg.Key),
		handlers.WithPrivateKey(s.cfg.PrivateKey),
		handlers.WithTrustedSubnet(s.cfg.TrustedSubnet),
	)

	s.httpServer = serviceHTTP.New(s.uc, h.GetRouter(),
		serviceHTTP.WithAddress(s.cfg.Addr),
	)
}

// runGraceDown Штатное завершение работы сервиса.
func (s *Service) runGraceDown() {
	// штатное завершение по сигналам: syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-sig

	// определяем время для штатного завершения работы сервера
	// необходимо на случай вечного ожидания закрытия всех подключений к серверу
	shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), serviceConst.GraceTimeout)
	defer shutdownCtxCancel()
	// принудительно завершаем работу по истечении срока s.graceTimeout
	go func() {
		<-shutdownCtx.Done()
		if shutdownCtx.Err() == context.DeadlineExceeded {
			log.Fatal("[ERROR] Graceful shutdown timed out! Forcing exit.")
		}
	}()

	// штатно завершим работу файлового хранилища и БД
	err := s.uc.Shutdown()
	if err != nil {
		log.Fatal("[ERROR] Shutdown error - ", err)
	}

	// Штатно завершаем работу HTTP-сервера не прерывая никаких активных подключений.
	// Завершение работы выполняется в порядке:
	// - закрытия всех открытых подключений;
	// - затем закрытия всех незанятых подключений;
	// - а затем бесконечного ожидания возврата подключений в режим ожидания;
	// - наконец, завершения работы.
	err = s.httpServer.Shutdown()
	if err != nil {
		log.Fatal("[ERROR] Server shutdown error - ", err)
	}
}

func (s *Service) Run() {
	go s.httpServer.Serve() // запускаем http-сервер

	s.runGraceDown()
}
