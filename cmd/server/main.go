package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/sergeysynergy/metricser/config"
	"github.com/sergeysynergy/metricser/internal/data/repository/filestore"
	"github.com/sergeysynergy/metricser/internal/data/repository/memory"
	"github.com/sergeysynergy/metricser/internal/data/repository/pgsql"
	"github.com/sergeysynergy/metricser/internal/handlers"
	"github.com/sergeysynergy/metricser/internal/httpserver"
	"github.com/sergeysynergy/metricser/internal/storage"
	"github.com/sergeysynergy/metricser/pkg/crypter"
	"github.com/sergeysynergy/metricser/pkg/utils"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const graceTimeout = 20 * time.Second

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
	repo := storage.Repo(memory.New())
	repoDB := pgsql.New(cfg.DatabaseDSN)
	if repoDB != nil {
		repo = storage.Repo(repoDB)
	}

	// Проверка на выполнение контракта интерфейса.
	var _ storage.FileRepo = new(filestore.FileStore)
	// Создадим файловое хранилище на базе Storage
	fileStorer := filestore.New(
		filestore.WithStoreFile(cfg.StoreFile),
		filestore.WithRestore(cfg.Restore),
		filestore.WithStoreInterval(cfg.StoreInterval),
	)

	uc := storage.New(
		storage.WithDBStorer(repo),
		storage.WithFileStorer(fileStorer),
	)

	// Подключим обработчики запросов.
	privateKey, err := crypter.OpenPrivate(cfg.CryptoKey)
	if err != nil {
		log.Println("[WARNING] Failed to get private key - ", err)
	}
	h := handlers.New(uc,
		handlers.WithKey(cfg.Key),
		handlers.WithPrivateKey(privateKey),
		handlers.WithTrustedSubnet(cfg.TrustedSubnet),
	)

	// Проинициализируем http-сервер с использованием ранее объявленных обработчиков и файлового хранилища.
	hs := httpserver.New(uc, h.GetRouter(),
		httpserver.WithAddress(cfg.Addr),
	)

	//go http.ListenAndServe(":8090", nil) // запускаем сервер для нужд профилирования

	go hs.Serve() // запускаем http-сервер

	graceDown(uc, hs)
}

// graceDown Штатное завершение работы сервиса.
func graceDown(uc storage.UseCase, hs *httpserver.Server) {
	// штатное завершение по сигналам: syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-sig

	// определяем время для штатного завершения работы сервера
	// необходимо на случай вечного ожидания закрытия всех подключений к серверу
	shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), graceTimeout)
	defer shutdownCtxCancel()
	// принудительно завершаем работу по истечении срока s.graceTimeout
	go func() {
		<-shutdownCtx.Done()
		if shutdownCtx.Err() == context.DeadlineExceeded {
			log.Fatal("[ERROR] Graceful shutdown timed out! Forcing exit.")
		}
	}()

	// штатно завершим работу файлового хранилища и БД
	err := uc.Shutdown()
	if err != nil {
		log.Fatal("[ERROR] Shutdown error - ", err)
	}

	// Штатно завершаем работу HTTP-сервера не прерывая никаких активных подключений.
	// Завершение работы выполняется в порядке:
	// - закрытия всех открытых подключений;
	// - затем закрытия всех незанятых подключений;
	// - а затем бесконечного ожидания возврата подключений в режим ожидания;
	// - наконец, завершения работы.
	err = hs.Shutdown()
	if err != nil {
		log.Fatal("[ERROR] Server shutdown error - ", err)
	}
}
