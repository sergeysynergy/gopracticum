package filestore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/sergeysynergy/gopracticum/internal/storage"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

var (
	ErrEmptyFilename = errors.New("filestore: empty filename")
)

type FileStore struct {
	*storage.Storage
	ctx           context.Context
	cancel        context.CancelFunc
	file          *os.File
	encoder       *json.Encoder
	storeFile     string        // имя файла, где хранятся значения метрик (пустое значение — отключает функцию записи на диск)
	storeInterval time.Duration // интервал периодического сохранения метрик на диск, 0 — делает запись синхронной
	restore       bool
	removeBroken  bool
}

type Options func(fs *FileStore)

func New(opts ...Options) storage.RepoStorer {
	const (
		defaultStoreFile     = "/tmp/devops-metrics-db.json"
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

	// создаём storage, если он не был проинициализирован через WithStorage
	if fs.Storage == nil {
		fs.Storage = storage.New()
	}

	// проинициализируем файловое хранилище
	err := fs.init()
	if err != nil {
		log.Fatal("[FATAL] Filestore initialization failed - ", err)
	}

	return fs
}

func WithStorage(st *storage.Storage) Options {
	return func(fs *FileStore) {
		fs.Storage = st
	}
}

func WithRestore(restore bool) Options {
	return func(fs *FileStore) {
		fs.restore = restore
	}
}

func WithStoreFile(filename string) Options {
	return func(fs *FileStore) {
		fs.storeFile = filename
	}
}

func WithStoreInterval(interval time.Duration) Options {
	return func(fs *FileStore) {
		fs.storeInterval = interval
	}
}

func (fs *FileStore) init() error {
	if fs.storeFile == "" {
		return ErrEmptyFilename
	}

	err := fs.restoreMetrics()
	if err != nil {
		log.Printf("[WARNING] Failed to restore metrics from file '%s' - %s\n", fs.storeFile, err)
	}

	file, err := os.OpenFile(fs.storeFile, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return fmt.Errorf("failed to create new filestore - %s", err)
	}

	fs.file = file
	fs.encoder = json.NewEncoder(file)

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

func (fs *FileStore) restoreMetrics() error {
	if !fs.restore {
		return nil
	}

	data, err := ioutil.ReadFile(fs.storeFile)
	if err != nil {
		return fs.removeBrokenFile(err)
	}

	m := metrics.ProxyMetric{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return fs.removeBrokenFile(err)
	}

	if len(m.Gauges) == 0 && len(m.Counters) == 0 {
		err = fmt.Errorf("metrics not found in file '%s'", fs.storeFile)
		return fs.removeBrokenFile(err)
	}

	fs.BulkPutGauges(m.Gauges)
	fs.BulkPutCounters(m.Counters)

	log.Printf("Restored metrics from file '%s': gauges %d, counters %d", fs.storeFile, len(m.Gauges), len(m.Counters))
	return nil
}

func (fs *FileStore) writeMetrics() (int, error) {
	m := &metrics.ProxyMetric{
		Gauges:   fs.GetGauges(),
		Counters: fs.GetCounters(),
	}

	_, err := fs.file.Seek(0, 0)
	if err != nil {
		return 0, err
	}

	err = fs.encoder.Encode(&m)
	if err != nil {
		return 0, err
	}

	log.Printf("Written metrics to file '%s': gauges %d, counters %d", fs.storeFile, len(m.Gauges), len(m.Counters))
	return len(m.Gauges) + len(m.Counters), nil
}

func (fs *FileStore) WriteTicker() error {
	// тикер должен работать только когда задано имя файла
	if fs.storeFile == "" {
		return ErrEmptyFilename
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
				_, err := fs.WriteMetrics()
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

// WriteMetrics Записываем метрики в файл, сработает только если storeInterval равен 0.
func (fs *FileStore) WriteMetrics() (int, error) {
	if fs.storeFile != "" && fs.storeInterval == 0 {
		number, err := fs.writeMetrics()
		if err != nil {
			return 0, fmt.Errorf("failed to store metrics in repository")
		}
		return number, nil
	}

	return 0, nil
}

func (fs *FileStore) Shutdown() error {
	defer fs.cancel()

	_, err := fs.writeMetrics()
	if err != nil {
		return err
	}

	err = fs.file.Close()
	if err != nil {
		return err
	}

	return nil
}
