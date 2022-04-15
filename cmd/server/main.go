package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/sergeysynergy/gopracticum/internal/filestore"
	"github.com/sergeysynergy/gopracticum/internal/handlers"
	"github.com/sergeysynergy/gopracticum/internal/httpserver"
	"github.com/sergeysynergy/gopracticum/internal/storage"
	"log"
	"os"
	"time"
)

type envConfig struct {
	Addr          string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
}

type config struct {
	addr          string
	restore       bool
	storeInterval time.Duration
	storeFile     string
}

func main() {
	envCfg := new(envConfig)
	err := env.Parse(envCfg)
	if err != nil {
		log.Fatalln(err)
	}

	cfg := new(config)
	flag.StringVar(&cfg.addr, "a", "127.0.0.1:8080", "address to listen on")
	flag.BoolVar(&cfg.restore, "r", true, "restore metrics from file")
	flag.DurationVar(&cfg.storeInterval, "i", 300*time.Second, "write metrics to file interval")
	flag.StringVar(&cfg.storeFile, "f", "/tmp/devops-metrics-db.json", "file to store metrics")
	flag.Parse()

	if _, ok := os.LookupEnv("ADDRESS"); ok {
		cfg.addr = envCfg.Addr
	}
	if _, ok := os.LookupEnv("RESTORE"); ok {
		cfg.restore = envCfg.Restore
	}
	if _, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		cfg.storeInterval = envCfg.StoreInterval
	}
	if _, ok := os.LookupEnv("STORE_FILE"); ok {
		cfg.storeFile = envCfg.StoreFile
	}
	fmt.Println(":: cfg", cfg)

	st := storage.New()
	fs := filestore.New(st,
		filestore.WithRestore(cfg.restore),
		filestore.WithStoreFile(cfg.storeFile),
		filestore.WithStoreInterval(cfg.storeInterval),
	)
	h := handlers.New(
		handlers.WithStorage(st),
		handlers.WithFileStore(fs),
	)

	s := httpserver.New(
		h.GetRouter(),
		fs,
		httpserver.WithAddress(cfg.addr),
	)

	s.Serve()
}
