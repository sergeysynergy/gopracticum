package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
	"time"

	"github.com/sergeysynergy/gopracticum/internal/agent"
)

type Config struct {
	Addr           string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	Key            string        `env:"KEY"`
}

func main() {
	cfg := new(Config)
	flag.StringVar(&cfg.Addr, "a", "127.0.0.1:8080", "server address")
	flag.DurationVar(&cfg.ReportInterval, "r", 10*time.Second, "interval for sending metrics to the server")
	flag.DurationVar(&cfg.PollInterval, "p", 2*time.Second, "update metrics interval")
	flag.StringVar(&cfg.Key, "k", "", "sign key")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	// создадим агента по сбору и отправке метрик
	// в качестве метрик выступают различные системные характеристики машины, на которой запущен агент
	a := agent.New(
		agent.WithAddress(cfg.Addr),
		agent.WithReportInterval(cfg.ReportInterval),
		agent.WithPollInterval(cfg.PollInterval),
		agent.WithKey(cfg.Key),
	)

	go http.ListenAndServe(":8091", nil) // запускаем сервер для нужд профилирования

	a.Run()
}
