// Package storage Пакет предназначен для реализации слоя бизнес-логики работы с метриками,
// а так же записи и извлечения их значений с использованием как БД, так и файлового хранилища.
package storage

import (
	"context"
	"errors"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
	"time"
)

var (
	ErrNotImplemented = errors.New("storage: metric not implemented")
	ErrNotFound       = errors.New("storage: metric not found")
)

// Storage Описывает логику работы с хранилищем метрик; кэширует значения всех метрик в памяти;
// сохраняет/извлекает метрики в БД.
type Storage struct {
	ctx           context.Context
	cancel        context.CancelFunc
	restore       bool
	storeInterval time.Duration // Интервал периодического сохранения метрик на диск, 0 — делает запись синхронной.

	repoDB   RepoDB
	repoFile RepoFile
}

// Проверим, что Storage соблюдает контракты интерфейса UseCases
var _ UseCase = new(Storage)

type Options func(storage *Storage)

// New Создаёт новый объект хранилища метрик Storage.
func New(repoDB RepoDB, repoFile RepoFile, opts ...Options) *Storage {
	const (
		defaultRestore       = false
		defaultStoreInterval = 300 * time.Second
	)

	ctx, cancel := context.WithCancel(context.Background())

	s := &Storage{
		ctx:           ctx,
		cancel:        cancel,
		restore:       defaultRestore,
		storeInterval: defaultStoreInterval,

		repoDB:   repoDB,
		repoFile: repoFile,
	}
	for _, opt := range opts {
		opt(s)
	}

	s.init()

	return s
}

// WithStoreInterval Определяет интервал сохранения метрик в файл.
func WithStoreInterval(interval time.Duration) Options {
	return func(s *Storage) {
		s.storeInterval = interval
	}
}

// WithRestore Определяет флаг, нужно ли восстанавливать при запуске значения метрик из файла.
func WithRestore(restore bool) Options {
	return func(s *Storage) {
		s.restore = restore
	}
}

// WithGauges Использует переданные значения gauge-метрик.
func WithGauges(gauges map[string]metrics.Gauge) Options {
	return func(s *Storage) {
		prm := metrics.NewProxyMetrics()
		prm.Gauges = gauges
		s.PutMetrics(prm)
	}
}

// WithCounters Использует переданные значения counter-метрик.
func WithCounters(counters map[string]metrics.Counter) Options {
	return func(s *Storage) {
		prm := metrics.NewProxyMetrics()
		prm.Counters = counters
		s.PutMetrics(prm)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *Storage) init() {
	err := s.restoreMetrics()
	if err != nil {
		log.Printf("[WARNING] Failed to restore metrics from file - %s", err)
	}
}
