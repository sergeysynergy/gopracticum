// Package filestore Пакет предназначен для записи/извлечения значений метрик в файл.
package filestore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sergeysynergy/metricser/internal/data/repository/memory"
	"log"
	"os"
	"time"

	metricserErrors "github.com/sergeysynergy/metricser/internal/errors"
	"github.com/sergeysynergy/metricser/internal/storage"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

// FileStore содержит реализацию репозитория работы с БД, контекст выполнения;
// реализует возможность записи и извлечения значений всех метрик из файла.
type FileStore struct {
	repo          storage.Repo
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
		repo:          memory.New(),
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

// Init производит инициализацию файлового хранилища.
func (fs *FileStore) init() error {
	if fs.storeFile == "" {
		return metricserErrors.EmptyFilename
	}

	err := fs.restoreMetrics()
	if err != nil {
		log.Printf("[WARNING] Failed to restore metrics from file '%s' - %s\n", fs.storeFile, err)
	}

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

// restoreMetrics Считывает все метрики из файла.
func (fs *FileStore) restoreMetrics() error {
	if !fs.restore {
		return nil
	}

	data, err := os.ReadFile(fs.storeFile)
	if err != nil {
		return fs.removeBrokenFile(err)
	}

	m := metrics.ProxyMetrics{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return fs.removeBrokenFile(err)
	}

	if len(m.Gauges) == 0 && len(m.Counters) == 0 {
		err = fmt.Errorf("metrics not found in file '%s'", fs.storeFile)
		return fs.removeBrokenFile(err)
	}

	log.Println("Read metrics from file:", string(data))

	err = fs.repo.Restore(&metrics.ProxyMetrics{Gauges: m.Gauges, Counters: m.Counters})
	if err != nil {
		return err
	}

	log.Printf("Restored metrics from file '%s': gauges %d, counters %d", fs.storeFile, len(m.Gauges), len(m.Counters))
	return nil
}

// writeMetrics Записывает показатели всех метрик в файл в JSON-формате.
func (fs *FileStore) writeMetrics() error {
	prm, _ := fs.repo.GetMetrics()

	data, err := json.Marshal(&prm)
	if err != nil {
		return err
	}

	err = os.WriteFile(fs.storeFile, data, 0777)
	if err != nil {
		return err
	}

	//log.Printf("written metrics to file '%s': gauges %d, counters %d", fs.storeFile, len(prm.Gauges), len(prm.Counters))
	//return len(prm.Gauges) + len(prm.Counters), nil
	return nil
}

// WriteTicker Асинхронно записывает метрики в файл с определённым интервалом.
func (fs *FileStore) WriteTicker() error {
	// тикер должен работать только когда задано имя файла
	if fs.storeFile == "" {
		return metricserErrors.EmptyFilename
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
				err := fs.WriteMetrics()
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

// WriteMetrics Записывает метрики в файл, сработает только если storeInterval равен 0.
func (fs *FileStore) WriteMetrics() error {
	if fs.storeFile != "" && fs.storeInterval == 0 {
		err := fs.writeMetrics()
		if err != nil {
			return fmt.Errorf("failed to store metrics in repository")
		}
		return nil
	}

	return nil
}

// Shutdown Штатно завершает работу файлового хранилища, сохраняя перед выходом значения метрик в файл.
func (fs *FileStore) Shutdown() error {
	defer fs.cancel()

	err := fs.writeMetrics()
	if err != nil {
		return err
	}

	return nil
}
