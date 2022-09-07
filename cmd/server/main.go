package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"log"
	"os"
	"strings"

	"github.com/sergeysynergy/metricser/config"
	"github.com/sergeysynergy/metricser/internal/data/repository/pgsql"
	"github.com/sergeysynergy/metricser/internal/filestore"
	"github.com/sergeysynergy/metricser/internal/handlers"
	"github.com/sergeysynergy/metricser/internal/httpserver"
	"github.com/sergeysynergy/metricser/pkg/crypter"
	"github.com/sergeysynergy/metricser/pkg/utils"
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

	// Получим путь к файлу из аргументов или переменной окружения.
	cfgFile, ok := os.LookupEnv("CONFIG")
	if !ok {
		for k, v := range os.Args[1:] {
			if v == "-c" && len(os.Args) > k+2 {
				cfgFile = os.Args[k+2]
			}
			if v == "-config" && len(os.Args) > k+2 {
				cfgFile = os.Args[k+2]
			}
			if strings.HasPrefix(v, "-c=") {
				cfgFile = os.Args[k+1][3:]
			}
			if strings.HasPrefix(v, "-config=") {
				cfgFile = os.Args[k+1][8:]
			}
		}
	}

	cfg := config.New()
	var err error

	// Загрузим конфигурацию из файла - самый низкий приоритет.
	if cfgFile != "" {
		log.Println("[INFO] Using config file:", cfgFile)
		cfg, err = config.LoadFromFile(cfgFile)
		if err != nil {
			log.Fatalln("[FATAL] Failed to load configuration -", err)
		}
	}

	// Перезапишем значения конфига значениями флагов, если те были переданы - средний приоритет.
	flag.StringVar(&cfgFile, "c", cfg.ConfigFile, "path to file with public key")
	flag.StringVar(&cfgFile, "config", cfg.ConfigFile, "path to file with public key")
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "address to listen on")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Postgres DSN")
	flag.StringVar(&cfg.StoreFile, "f", cfg.StoreFile, "file to store metrics")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "sign key")
	flag.DurationVar(&cfg.StoreInterval.Duration, "i", cfg.StoreInterval.Duration, "interval for saving to file")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "restore metrics from file")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "path to file with public key")
	flag.Parse()

	// Перезапишем значения конфига переменными окружения - самый главный приоритет.
	err = env.Parse(cfg)
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
