// Package filestore Пакет предназначен для записи/извлечения значений метрик в файл.
package filestore

import (
	"context"
	"fmt"
	"github.com/sergeysynergy/metricser/internal/service/data/repository/memory"
	serviceErrors "github.com/sergeysynergy/metricser/internal/service/errors"
	"github.com/sergeysynergy/metricser/internal/service/storage"
	"log"
	"os"
	"time"
)

// FileStore содержит реализацию репозитория работы с БД, контекст выполнения;
// реализует возможность записи и извлечения значений всех метрик из файла.
type FileStore struct {
	repo          storage.Repo
	ctx           context.Context
	cancel        context.CancelFunc
	storeFile     string        // Имя файла, где хранятся значения метрик (пустое значение — отключает функцию записи на диск).
	storeInterval time.Duration // Интервал периодического сохранения метрик на диск, 0 — делает запись синхронной.
	removeBroken  bool
}

type Options func(fs *FileStore)

// New Создаёт новый объект файлового хранилища FileStorer.
func New(opts ...Options) *FileStore {
	const (
		defaultStoreFile     = "/tmp/devops-metrics-pgsql.json"
		defaultStoreInterval = 300 * time.Second
		defaultRemoveBroken  = false
	)

	ctx, cancel := context.WithCancel(context.Background())

	fs := &FileStore{
		repo:          memory.New(),
		ctx:           ctx,
		cancel:        cancel,
		storeFile:     defaultStoreFile,
		storeInterval: defaultStoreInterval,
		removeBroken:  defaultRemoveBroken,
	}
	for _, opt := range opts {
		opt(fs)
	}

	// вернём nil в случае пустого имени файла
	if fs.storeFile == "" {
		return nil
	}

	// проинициализируем файловое хранилище
	err := fs.init()
	if err != nil {
		log.Fatal("[FATAL] File store initialization failed - ", err)
	}

	return fs
}

// WithStorer Использует переданный репозиторий.
func WithStorer(repo storage.Repo) Options {
	return func(fs *FileStore) {
		if repo != nil {
			fs.repo = repo
		}
	}
}

// WithStoreFile Использует переданный путь к файлу.
func WithStoreFile(filename string) Options {
	return func(fs *FileStore) {
		fs.storeFile = filename
	}
}

// WithStoreInterval Определяет интервал сохранения метрик в файл.
func WithStoreInterval(interval time.Duration) Options {
	return func(fs *FileStore) {
		fs.storeInterval = interval
	}
}

// Init производит инициализацию файлового хранилища.
func (fs *FileStore) init() error {
	//if fs.storeFile == "" {
	//	return serviceErrors.ErrEmptyFilestoreName
	//}
	//
	//err := fs.restoreMetrics()
	//if err != nil {
	//	log.Printf("[WARNING] Failed to restore metrics from file '%s' - %s\n", fs.storeFile, err)
	//}

	return nil
}

func (fs *FileStore) removeBrokenFile(err error) error {
	if !fs.removeBroken {
		return err
	}

	errRm := os.Remove(fs.storeFile)
	if errRm != nil {
		return errRm
	}

	return err
}

// WriteTicker Асинхронно записывает метрики в файл с определённым интервалом.
func (fs *FileStore) WriteTicker() error {
	// тикер должен работать только когда задано имя файла
	if fs.storeFile == "" {
		return serviceErrors.ErrEmptyFilestoreName
	}
	// ... и storeInterval больше нуля
	if fs.storeInterval == 0 {
		return fmt.Errorf("store interval should be > 0 to start WriteTicker routine")
	}

	go func() {
		ticker := time.NewTicker(fs.storeInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				prm, _ := fs.repo.GetMetrics()
				err := fs.JustWriteMetrics(prm)
				if err != nil {
					log.Println("[ERROR] Failed to write metrics to disk -", err)
				}
			case <-fs.ctx.Done():
				return
			}
		}
	}()

	return nil
}
