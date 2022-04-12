package main

import (
	"github.com/caarlos0/env/v6"
	"log"
	"os"
	"time"

	"github.com/sergeysynergy/gopracticum/internal/filestore"
	"github.com/sergeysynergy/gopracticum/internal/handlers"
	"github.com/sergeysynergy/gopracticum/internal/httpserver"
	"github.com/sergeysynergy/gopracticum/internal/storage"
)

type Config struct {
	Addr          string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	Restore       bool          `env:"RESTORE"`
}

func main() {
	var cfg = Config{
		Addr:          "127.0.0.1:8080",
		StoreInterval: 300 * time.Second,
		Restore:       true,
	}

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalln(err)
	}

	st := storage.New()
	h := handlers.New(handlers.WithStorage(st))

	// в случае отсутствия переменной окружения STORE_FILE создаём объёкт с именем файла по умолчанию
	var fs *filestore.FileStore
	storeFile, ok := os.LookupEnv("STORE_FILE")
	if !ok {
		fs = filestore.New(st,
			filestore.WithRestore(cfg.Restore),
			filestore.WithStoreInterval(cfg.StoreInterval),
		)
	} else {
		fs = filestore.New(st,
			filestore.WithRestore(cfg.Restore),
			filestore.WithStoreFile(storeFile),
			filestore.WithStoreInterval(cfg.StoreInterval),
		)
	}

	s := httpserver.New(
		h.GetRouter(),
		fs,
		httpserver.WithAddress(cfg.Addr),
	)

	s.Serve()
}
