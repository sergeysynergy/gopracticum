package filestore

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/sergeysynergy/gopracticum/internal/storage"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

type storeMetrics struct {
	Gauges   map[string]metrics.Gauge
	Counters map[string]metrics.Counter
}

type FileStore struct {
	ctx           context.Context
	cancel        context.CancelFunc
	storage       storage.Storer
	file          *os.File
	encoder       *json.Encoder
	storeFile     string        // имя файла, где хранятся значения метрик (пустое значение — отключает функцию записи на диск)
	storeInterval time.Duration // интервал периодического сохранения метрик на диск, 0 — делает запись синхронной
	restore       bool
}

type Option func(fs *FileStore)

func New(st storage.Storer, opts ...Option) *FileStore {
	const (
		defaultRestore       = true
		defaultStoreFile     = "/tmp/devops-metrics-db.json"
		defaultStoreInterval = 300 * time.Second
	)

	ctx, cancel := context.WithCancel(context.Background())

	fs := &FileStore{
		ctx:           ctx,
		cancel:        cancel,
		storage:       st,
		storeFile:     defaultStoreFile,
		storeInterval: defaultStoreInterval,
		restore:       defaultRestore,
	}
	for _, opt := range opts {
		opt(fs)
	}

	// восстановим из файла в хранилище значения метрик, если restore = true
	err := fs.restoreMetrics()
	if err != nil {
		log.Printf("[ERROR] Failed to restore metrics from file '%s' - %s\n", fs.storeFile, err)
	}

	// если специально (через WithStoreFile) задано пустое имя файла, возвращаем nil вместо объекта
	if fs.storeFile == "" {
		return nil
	}

	file, err := os.OpenFile(fs.storeFile, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Println("[ERROR] Failed to create new file store - ", err)
		return nil
	}

	fs.file = file
	fs.encoder = json.NewEncoder(file)

	return fs
}

func WithRestore(restore bool) Option {
	return func(fs *FileStore) {
		fs.restore = restore
	}
}

func WithStoreFile(filename string) Option {
	return func(fs *FileStore) {
		fs.storeFile = filename
	}
}

func WithStoreInterval(interval time.Duration) Option {
	return func(fs *FileStore) {
		fs.storeInterval = interval
	}
}

func (fs *FileStore) restoreMetrics() error {
	if !fs.restore {
		return nil
	}

	data, err := ioutil.ReadFile(fs.storeFile)
	if err != nil {
		os.Remove(fs.storeFile)
		return err
	}

	m := storeMetrics{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		os.Remove(fs.storeFile)
		return err
	}

	if len(m.Gauges) == 0 && len(m.Counters) == 0 {
		os.Remove(fs.storeFile)
		return fmt.Errorf("metrics not found in file '%s'", fs.storeFile)
	}

	fs.storage.BulkPutGauges(m.Gauges)
	fs.storage.BulkPutCounters(m.Counters)

	log.Printf("Restored metrics from file '%s': gauges %d, counters %d", fs.storeFile, len(m.Gauges), len(m.Counters))
	return nil
}

func (fs *FileStore) WriteMetrics() error {
	m := &storeMetrics{
		Gauges:   fs.storage.GetGauges(),
		Counters: fs.storage.GetCounters(),
	}

	_, err := fs.file.Seek(0, 0)
	if err != nil {
		return err
	}

	err = fs.encoder.Encode(&m)
	if err != nil {
		return err
	}

	log.Printf("Written metrics to file '%s': gauges %d, counters %d", fs.storeFile, len(m.Gauges), len(m.Counters))
	return nil
}

func (fs *FileStore) Shutdown() error {
	defer fs.cancel()

	err := fs.WriteMetrics()
	if err != nil {
		return err
	}

	err = fs.file.Close()
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileStore) WriteTicker() {
	if fs.storeInterval == 0 {
		err := fs.WriteMetrics()
		if err != nil {
			log.Println("[ERROR] Failed to write metrics to disk -", err)
		}

		return
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
}
