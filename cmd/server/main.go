package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
	"time"

	"github.com/sergeysynergy/gopracticum/internal/filestore"
	"github.com/sergeysynergy/gopracticum/internal/handlers"
	"github.com/sergeysynergy/gopracticum/internal/httpserver"
)

type config struct {
	Addr          string        `env:"ADDRESS"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	Key           string        `env:"KEY"`
}

func main() {
	cfg := new(config)
	flag.StringVar(&cfg.Addr, "a", "127.0.0.1:8080", "address to listen on")
	flag.StringVar(&cfg.StoreFile, "f", "/tmp/devops-metrics-db.json", "file to store metrics")
	flag.BoolVar(&cfg.Restore, "r", true, "restore metrics from file")
	flag.DurationVar(&cfg.StoreInterval, "i", 300*time.Second, "interval for saving metrics in repository")
	flag.StringVar(&cfg.Key, "k", "", "sign key")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	// объявим новое хранилище, которое реализует интерфейс Storer
	//st := storage.New()

	// создадим файловое хранилище на базе storage
	repoStorer := filestore.New(
		filestore.WithStoreFile(cfg.StoreFile),
		filestore.WithRestore(cfg.Restore),
		filestore.WithStoreInterval(cfg.StoreInterval),
	)

	// подключим обработчики запросов, которые используют storage и fileStore
	h := handlers.New(
		handlers.WithRepoStorer(repoStorer),
		handlers.WithKey(cfg.Key),
	)

	// проиницилизируем сервер с использованием ранее объявленных обработчиков и файлового хранилища
	s := httpserver.New(
		h.GetRouter(),
		repoStorer,
		httpserver.WithAddress(cfg.Addr),
	)

	s.Serve()
}
