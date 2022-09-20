package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"log"

	"github.com/sergeysynergy/metricser/config"
	"github.com/sergeysynergy/metricser/internal/service"
	"github.com/sergeysynergy/metricser/internal/service/data/repository/memory"
	"github.com/sergeysynergy/metricser/internal/service/data/repository/pgsql"
	"github.com/sergeysynergy/metricser/internal/service/filestore"
	"github.com/sergeysynergy/metricser/internal/service/storage"
	"github.com/sergeysynergy/metricser/pkg/utils"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	var err error

	// Выведем номер версии, сборки и комит, если доступны.
	// Для задания переменных рекомендуется использовать опции линковщика, например:
	// go run -ldflags "-X main.buildVersion=v1.0.1" main.go
	fmt.Printf("Build version: %s\n", utils.CheckNA(buildVersion))
	fmt.Printf("Build date: %s\n", utils.CheckNA(buildDate))
	fmt.Printf("Build commint: %s\n", utils.CheckNA(buildCommit))

	// Получим конфиг: попытаемся загрузить его из файла.
	cfg := config.NewServerConf()

	// Перезапишем значения конфига значениями флагов, если те были переданы - средний приоритет.
	flag.StringVar(&cfg.ConfigFile, "c", cfg.ConfigFile, "path to file with public key")
	flag.StringVar(&cfg.ConfigFile, "config", cfg.ConfigFile, "path to file with public key")
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "address to listen on")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Postgres DSN")
	flag.StringVar(&cfg.StoreFile, "f", cfg.StoreFile, "file to store metrics")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "sign key")
	flag.DurationVar(&cfg.StoreInterval, "i", cfg.StoreInterval, "interval for saving to file")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "restore metrics from file")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "path to file with public key")
	flag.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "CIDR - Classless Inter-Domain Routing")
	flag.Parse()

	// Перезапишем значения конфига переменными окружения - самый главный приоритет.
	err = env.Parse(cfg)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("[DEBUG] Receive config: %#v\n", cfg)
	cfg.Init()

	// Проверка на выполнение контракта интерфейса.
	var _ storage.Repo = new(pgsql.Storage)

	// Получим реализацию репозитория для работы с БД.
	repo := storage.Repo(memory.New())
	repoDB := pgsql.New(cfg.DatabaseDSN)
	if repoDB != nil {
		repo = storage.Repo(repoDB)
	}

	// Проверка на выполнение контракта интерфейса.
	var _ storage.FileRepo = new(filestore.FileStore)
	// Создадим файловое хранилище на базе Storage
	fileStorer := filestore.New(
		filestore.WithStorer(repo),
		filestore.WithStoreFile(cfg.StoreFile),
		filestore.WithStoreInterval(cfg.StoreInterval),
	)

	uc := storage.New(
		storage.WithDBStorer(repo),
		storage.WithFileStorer(fileStorer),
		storage.WithRestore(cfg.Restore),
	)

	// Подключим обработчики запросов.

	srv := service.New(cfg, uc)
	srv.Run()
}
