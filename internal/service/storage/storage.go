// Package storage Пакет предназначен для реализации слоя бизнес-логики работы с метриками,
// а так же записи и извлечения их значений с использованием как БД, так и файлового хранилища.
package storage

import (
	"context"
	"github.com/sergeysynergy/metricser/internal/service/data/repository/memory"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
	"time"
)

// Storage Описывает логику работы с хранилищем метрик; кэширует значения всех метрик в памяти;
// сохраняет/извлекает метрики в БД.
type Storage struct {
	repo          Repo
	fileRepo      FileRepo
	ctx           context.Context
	cancel        context.CancelFunc
	restore       bool
	storeInterval time.Duration // Интервал периодического сохранения метрик на диск, 0 — делает запись синхронной.
}

type Option func(storage *Storage)

// New Создаёт новый объект хранилища метрик Storage.
func New(opts ...Option) *Storage {
	const defaultStoreInterval = 300 * time.Second

	ctx, cancel := context.WithCancel(context.Background())

	s := &Storage{
		ctx:           ctx,
		cancel:        cancel,
		repo:          memory.New(),
		fileRepo:      nil,
		storeInterval: defaultStoreInterval,
	}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func WithDBStorer(repo Repo) Option {
	return func(s *Storage) {
		if repo != nil {
			log.Println("[DEBUG] Database plugin connected")
			s.repo = repo
		}
	}
}

func WithFileStorer(fr FileRepo) Option {
	return func(s *Storage) {
		if fr != nil {
			log.Println("[DEBUG] File store plugin connected")
			s.fileRepo = fr
		} else {
			log.Println("[DEBUG] File store plugin is <nil>")
		}
	}
}

// WithGauges Использует переданные значения gauge-метрик.
func WithGauges(gauges map[string]metrics.Gauge) Option {
	return func(s *Storage) {
		prm := metrics.NewProxyMetrics()
		prm.Gauges = gauges
		err := s.repo.Restore(prm)
		if err != nil {
			log.Println("[ERROR] Failed to init storage with gauges")
		}
	}
}

// WithCounters Использует переданные значения counter-метрик.
func WithCounters(counters map[string]metrics.Counter) Option {
	return func(s *Storage) {
		//s.counters = counters
		prm := metrics.NewProxyMetrics()
		prm.Counters = counters
		err := s.repo.Restore(prm)
		if err != nil {
			log.Println("[ERROR] Failed to init storage with counters")
		}
	}
}

// WithRestore Определяет флаг, нужно ли восстанавливать при запуске значения метрик из файла.
func WithRestore(restore bool) Option {
	return func(s *Storage) {
		s.restore = restore
	}
}

// WithStoreInterval Определяет интервал сохранения метрик в файл.
func WithStoreInterval(interval time.Duration) Option {
	return func(s *Storage) {
		s.storeInterval = interval
	}
}

func (s *Storage) init() {
	if s.restore {
		err := s.snapShotRestore()
		if err != nil {
			log.Printf("[WARNING] Failed to restore metrics from filestore - %s\n", err)
		}
	}
}
