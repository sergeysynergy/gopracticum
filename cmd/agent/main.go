package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/sergeysynergy/gopracticum/internal/agent"
	"log"
	"time"
)

type Config struct {
	Addr           string        `env:"ADDRESS"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
}

func main() {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalln(err)
	}

	a := agent.New(
		agent.WithAddress(cfg.Addr),
		agent.WithPollInterval(cfg.PollInterval),
		agent.WithReportInterval(cfg.ReportInterval),
	)

	a.Run()
}
