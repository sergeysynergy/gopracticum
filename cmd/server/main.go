package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/sergeysynergy/metricser/config"
	"github.com/sergeysynergy/metricser/internal/data/repository/pgsql"
	"github.com/sergeysynergy/metricser/internal/filestore"
	"github.com/sergeysynergy/metricser/internal/handlers"
	"github.com/sergeysynergy/metricser/internal/httpserver"
	"github.com/sergeysynergy/metricser/internal/storage"
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
	log.Printf("Receive config: %#v\n", cfg)

	// Проверка на выполнение контракта интерфейса.
	var _ storage.Repo = new(pgsql.Storage)
	// Получим реализацию репозитория для работы с БД.
	dbStorer := pgsql.New(cfg.DatabaseDSN)

	// Проверка на выполнение контракта интерфейса.
	var _ storage.FileRepo = new(filestore.FileStore)
	// Создадим файловое хранилище на базе Storage
	fileStorer := filestore.New(
		filestore.WithStorer(dbStorer),
		filestore.WithStoreFile(cfg.StoreFile),
		filestore.WithRestore(cfg.Restore),
		filestore.WithStoreInterval(cfg.StoreInterval),
	)

	uc := storage.New(dbStorer, fileStorer,
		storage.WithDBStorer(dbStorer),
		storage.WithFileStorer(fileStorer),
	)

	// Подключим обработчики запросов.
	privateKey, err := crypter.OpenPrivate(cfg.CryptoKey)
	if err != nil {
		log.Println("[WARNING] Failed to get private key - ", err)
	}
	h := handlers.New(uc,
		//handlers.WithFileStorer(fileStorer),
		//handlers.WithDBStorer(dbStorer),
		handlers.WithKey(cfg.Key),
		handlers.WithPrivateKey(privateKey),
		handlers.WithTrustedSubnet(cfg.TrustedSubnet),
	)

	// Проинициализируем сервер с использованием ранее объявленных обработчиков и файлового хранилища.
	s := httpserver.New(uc, h.GetRouter(),
		httpserver.WithAddress(cfg.Addr),
		//httpserver.WithFileStorer(fileStorer),
		//httpserver.WithDBStorer(dbStorer),
	)

	//go http.ListenAndServe(":8090", nil) // запускаем сервер для нужд профилирования

	s.Serve() // запускаем основной http-сервер
}
