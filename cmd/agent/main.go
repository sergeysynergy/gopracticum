package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/sergeysynergy/gopracticum/internal/agent"
	"log"
	"os"
	"time"
)

type Config struct {
}

type envConfig struct {
	Addr           string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}

type config struct {
	addr           string
	reportInterval time.Duration
	pollInterval   time.Duration
}

func main() {
	envCfg := new(envConfig)
	err := env.Parse(envCfg)
	if err != nil {
		log.Fatalln(err)
	}

	cfg := new(config)
	flag.StringVar(&cfg.addr, "a", "127.0.0.1:8080", "address to listen on")
	flag.DurationVar(&cfg.reportInterval, "r", 10*time.Second, "write metrics to file interval")
	flag.DurationVar(&cfg.pollInterval, "p", 2*time.Second, "write metrics to file interval")
	flag.Parse()

	if _, ok := os.LookupEnv("ADDRESS"); ok {
		cfg.addr = envCfg.Addr
	}
	if _, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		cfg.reportInterval = envCfg.ReportInterval
	}
	if _, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		cfg.pollInterval = envCfg.PollInterval
	}

	a := agent.New(
		agent.WithAddress(cfg.addr),
		agent.WithReportInterval(cfg.reportInterval),
		agent.WithPollInterval(cfg.pollInterval),
	)

	a.Run()
}
