package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/sergeysynergy/metricser/config"
	"github.com/sergeysynergy/metricser/pkg/crypter"
	"github.com/sergeysynergy/metricser/pkg/utils"
	"log"
	//_ "net/http/pprof" // подключаем пакет pprof
	"time"

	"github.com/sergeysynergy/metricser/internal/data/repository/pgsql"
	"github.com/sergeysynergy/metricser/internal/filestore"
	"github.com/sergeysynergy/metricser/internal/handlers"
	"github.com/sergeysynergy/metricser/internal/httpserver"
)

//type config struct {
//	Key        string `env:"KEY"`
//	ConfigFile string
//
//	Addr          string        `env:"ADDRESS" json:"address"`
//	StoreFile     string        `env:"STORE_FILE" json:"store_file"`
//	Restore       bool          `env:"RESTORE" json:"restore"`
//	StoreInterval time.Duration `env:"STORE_INTERVAL" json:"store_interval"`
//	DatabaseDSN   string        `env:"DATABASE_DSN" json:"database_dsn"`
//	CryptoKey     string        `env:"CRYPTO_KEY" json:"crypto_key"`
//}

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

	cfg := config.New()
	flag.StringVar(&cfg.Addr, "a", "127.0.0.1:8080", "address to listen on")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Postgres DSN")
	flag.StringVar(&cfg.StoreFile, "f", "/tmp/devops-metrics-pgsql.json", "file to store metrics")
	flag.StringVar(&cfg.Key, "k", "", "sign key")
	flag.DurationVar(&cfg.StoreInterval.Duration, "i", 300*time.Second, "interval for saving to file")
	flag.BoolVar(&cfg.Restore, "r", true, "restore metrics from file")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "path to file with public key")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Receive config: %#v\n", cfg)

	// Получим реализацию репозитория для работы с БД.
	dbStorer := pgsql.New(cfg.DatabaseDSN)

	// Создадим файловое хранилище на базе Storage
	fileStorer := filestore.New(
		filestore.WithStorer(dbStorer),
		filestore.WithStoreFile(cfg.StoreFile),
		filestore.WithRestore(cfg.Restore),
		filestore.WithStoreInterval(cfg.StoreInterval.Duration),
	)

	privateKey, err := crypter.OpenPrivate(cfg.CryptoKey)
	if err != nil {
		log.Println("[WARNING] Failed to get private key - ", err)
	}

	// Подключим обработчики запросов.
	h := handlers.New(
		handlers.WithFileStorer(fileStorer),
		handlers.WithDBStorer(dbStorer),
		handlers.WithKey(cfg.Key),
		handlers.WithPrivateKey(privateKey),
	)

	// Проинициализируем сервер с использованием ранее объявленных обработчиков и файлового хранилища.
	s := httpserver.New(h.GetRouter(),
		httpserver.WithAddress(cfg.Addr),
		httpserver.WithFileStorer(fileStorer),
		httpserver.WithDBStorer(dbStorer),
	)

	//go http.ListenAndServe(":8090", nil) // запускаем сервер для нужд профилирования

	s.Serve() // запускаем основной http-сервер
}
