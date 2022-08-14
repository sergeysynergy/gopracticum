package agent

import (
	"context"
	"github.com/go-resty/resty/v2"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sergeysynergy/metricser/internal/storage"
)

type Agent struct {
	client         *resty.Client
	storage        storage.Repo
	pollInterval   time.Duration
	reportInterval time.Duration
	protocol       string
	addr           string
	key            string
}

type Option func(agent *Agent)

func New(opts ...Option) *Agent {
	rand.Seed(time.Now().UTC().UnixNano())

	const (
		defaultReportInterval = 10 * time.Second // частота отправки метрик на сервер
		defaultPollInterval   = 2 * time.Second  // частота обновления метрик из пакета `runtime`
		defaultAddress        = "127.0.0.1:8080"
		defaultProtocol       = "http://"
		defaultTimeout        = 4 * time.Second
	)

	a := &Agent{
		client:         resty.New(),
		storage:        storage.New(),
		pollInterval:   defaultPollInterval,
		reportInterval: defaultReportInterval,
		protocol:       defaultProtocol,
		addr:           defaultAddress,
	}
	a.client.SetTimeout(defaultTimeout)

	// Применяем в цикле каждую опцию
	for _, opt := range opts {
		// вызываем функцию, предоставляющую экземпляр *Agent в качестве аргумента
		opt(a)
	}

	// вернуть измененный экземпляр Server
	return a
}

func WithAddress(addr string) Option {
	return func(a *Agent) {
		if addr != "" {
			a.addr = addr
		}
	}
}

func WithPollInterval(duration time.Duration) Option {
	return func(a *Agent) {
		if duration > 0 {
			a.pollInterval = duration
		}
	}
}

func WithReportInterval(duration time.Duration) Option {
	return func(a *Agent) {
		if duration > 0 {
			a.reportInterval = duration
		}
	}
}

func WithKey(key string) Option {
	return func(a *Agent) {
		a.key = key
	}
}

func (a *Agent) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	// Функцию cancel нужно обязательно выполнить в коде, иначе сборщик мусора не удалит созданный дочерний контекст
	// и произойдёт утечка памяти.
	defer cancel()

	go a.pollTicker(ctx)
	go a.gopsutilTicker(ctx)
	go a.reportTicker(ctx)

	// Агент должен штатно завершаться по сигналам: syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-c
	log.Println("Получен сигнал завершения работы:", sig)
	log.Println("Работа агента штатно завершена")
}

func (a *Agent) handleError(err error) {
	log.Println("Ошибка -", err)
}
