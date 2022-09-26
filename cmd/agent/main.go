package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/sergeysynergy/metricser/config"
	"github.com/sergeysynergy/metricser/internal/agent"
	"github.com/sergeysynergy/metricser/pkg/crypter"
	"github.com/sergeysynergy/metricser/pkg/utils"
	"log"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	// Выведем номер версии, сборки и комит, если доступны.
	// Для задания переменных рекомендуется использовать опции линковщика, например:
	// go run -ldflags "-X main.buildVersion=v1.0.1" main.go
	fmt.Printf("Build version: %s\n", utils.CheckNA(buildVersion))
	fmt.Printf("Build date: %s\n", utils.CheckNA(buildDate))
	fmt.Printf("Build commint: %s\n", utils.CheckNA(buildCommit))

	cfg := config.NewAgentConf()
	flag.StringVar(&cfg.ConfigFile, "c", cfg.ConfigFile, "path to file with public key")
	flag.StringVar(&cfg.ConfigFile, "config", cfg.ConfigFile, "path to file with public key")
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "server address")
	flag.DurationVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "interval for sending metrics to the server")
	flag.DurationVar(&cfg.PollInterval, "p", cfg.PollInterval, "update metrics interval")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "sign key")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "path to file with public key")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	pubKey, err := crypter.OpenPublic(cfg.CryptoKey)
	if err != nil {
		log.Println("[WARNING] Failed to get public key -", err)
	}

	// создадим агента по сбору и отправке метрик
	// в качестве метрик выступают различные системные характеристики машины, на которой запущен агент
	a := agent.New(
		agent.WithAddress(cfg.Addr),
		agent.WithReportInterval(cfg.ReportInterval),
		agent.WithPollInterval(cfg.PollInterval),
		agent.WithKey(cfg.Key),
		agent.WithPublicKey(pubKey),
	)

	a.Run()
}
