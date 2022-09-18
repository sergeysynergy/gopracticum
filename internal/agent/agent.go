// Package agent Пакет реализует клиента для сборки и отправки метрик на сервер сбора и хранения этих метрик.
package agent

import (
	"crypto/rsa"
	"github.com/go-resty/resty/v2"
	"github.com/sergeysynergy/metricser/internal/data/repository/memory"
	"math/rand"
	"time"

	"github.com/sergeysynergy/metricser/internal/domain/storage"
)

type Agent struct {
	client         *resty.Client
	repo           storage.RepoDB
	pollInterval   time.Duration
	reportInterval time.Duration
	protocol       string
	addr           string
	key            string
	publicKey      *rsa.PublicKey
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
		repo:           memory.New(),
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

func WithPublicKey(key *rsa.PublicKey) Option {
	return func(a *Agent) {
		a.publicKey = key
	}
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
