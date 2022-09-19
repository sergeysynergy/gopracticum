// Package filestore Пакет предназначен для записи/извлечения значений метрик в файл.
package filestore

import (
	"context"
	"fmt"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
	"os"
	"time"

	metricserErrors "github.com/sergeysynergy/metricser/internal/errors"
)

// FileStore содержит реализацию репозитория работы с БД, контекст выполнения;
// реализует возможность записи и извлечения значений всех метрик из файла.
type FileStore struct {
	ctx           context.Context
	cancel        context.CancelFunc
	storeFile     string        // Имя файла, где хранятся значения метрик (пустое значение — отключает функцию записи на диск).
	storeInterval time.Duration // Интервал периодического сохранения метрик на диск, 0 — делает запись синхронной.
	restore       bool
	removeBroken  bool
}

type Options func(fs *FileStore)

// New Создаёт новый объект файлового хранилища FileStorer.
func New(opts ...Options) *FileStore {
	const (
		defaultStoreFile     = "/tmp/devops-metrics-pgsql.json"
		defaultRestore       = false
		defaultStoreInterval = 300 * time.Second
		defaultRemoveBroken  = false
	)

	ctx, cancel := context.WithCancel(context.Background())

	fs := &FileStore{
		ctx:           ctx,
		cancel:        cancel,
		storeFile:     defaultStoreFile,
		restore:       defaultRestore,
		storeInterval: defaultStoreInterval,
		removeBroken:  defaultRemoveBroken,
	}
	for _, opt := range opts {
		opt(fs)
	}

	// вернём nil в случае пустого имени файла
	if fs.storeFile == "" {
		log.Fatal("[FATAL] File store initialization failed - ", metricserErrors.ErrEmptyFilename)
	}

	return fs
}

// WithRestore Определяет флаг, нужно ли восстанавливать при запуске значения метрик из файла.
func WithRestore(restore bool) Options {
	return func(fs *FileStore) {
		fs.restore = restore
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
func (fs *FileStore) WriteTicker(prm *metrics.ProxyMetrics) error {
	// тикер должен работать только когда задано имя файла
	if fs.storeFile == "" {
		return metricserErrors.ErrEmptyFilename
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
				if fs.storeFile != "" && fs.storeInterval == 0 {
					err := fs.WriteMetrics(prm)
					if err != nil {
						log.Println("[ERROR] Failed to write metrics to disk -", err)
					}
				}
			case <-fs.ctx.Done():
				return
			}
		}
	}()

	return nil
}
