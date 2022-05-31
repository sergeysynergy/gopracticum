package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
	"time"

	"github.com/sergeysynergy/metricser/internal/db"
	"github.com/sergeysynergy/metricser/internal/filestore"
	"github.com/sergeysynergy/metricser/internal/handlers"
	"github.com/sergeysynergy/metricser/internal/httpserver"
)

type config struct {
	Addr          string        `env:"ADDRESS"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	Key           string        `env:"KEY"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`
}

func main() {
	cfg := new(config)
	flag.StringVar(&cfg.Addr, "a", "127.0.0.1:8080", "address to listen on")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Postgres DSN")
	flag.StringVar(&cfg.StoreFile, "f", "/tmp/devops-metrics-db.json", "file to store metrics")
	flag.StringVar(&cfg.Key, "k", "", "sign key")
	flag.DurationVar(&cfg.StoreInterval, "i", 300*time.Second, "interval for saving to file")
	flag.BoolVar(&cfg.Restore, "r", true, "restore metrics from file")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Receive config: %#v\n", cfg)

	// создадим хранилище с использование базы данных на базе Storage
	dbStorer := db.New(cfg.DatabaseDSN)

	// создадим файловое хранилище на базе Storage
	fileStorer := filestore.New(
		filestore.WithStorer(dbStorer),
		filestore.WithStoreFile(cfg.StoreFile),
		filestore.WithRestore(cfg.Restore),
		filestore.WithStoreInterval(cfg.StoreInterval),
	)

	// подключим обработчики запросов, которые используют storage и fileStore
	h := handlers.New(
		handlers.WithFileStorer(fileStorer),
		handlers.WithDBStorer(dbStorer),
		handlers.WithKey(cfg.Key),
	)

	// проиницилизируем сервер с использованием ранее объявленных обработчиков и файлового хранилища
	s := httpserver.New(h.GetRouter(),
		httpserver.WithAddress(cfg.Addr),
		httpserver.WithFileStorer(fileStorer),
		httpserver.WithDBStorer(dbStorer),
	)

	s.Serve()
}
