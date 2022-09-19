// Package storage Пакет предназначен для реализации слоя бизнес-логики работы с метриками,
// а так же записи и извлечения их значений с использованием как БД, так и файлового хранилища.
package storage

import (
	"errors"
	"github.com/sergeysynergy/metricser/internal/data/repository/memory"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
)

var (
	ErrNotImplemented = errors.New("metrics not implemented")
	ErrNotFound       = errors.New("metrics not found")
)

// Storage Описывает логику работы с хранилищем метрик; кэширует значения всех метрик в памяти;
// сохраняет/извлекает метрики в БД.
type Storage struct {
	repo     Repo
	fileRepo FileRepo
}

type Option func(storage *Storage)

// New Создаёт новый объект хранилища метрик Storage.
func New(opts ...Option) *Storage {
	s := &Storage{
		repo:     memory.New(),
		fileRepo: nil,
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
		//s.gauges = gauges
		prm := metrics.NewProxyMetrics()
		prm.Gauges = gauges
		s.repo.Restore(prm)
	}
}

// WithCounters Использует переданные значения counter-метрик.
func WithCounters(counters map[string]metrics.Counter) Option {
	return func(s *Storage) {
		//s.counters = counters
		prm := metrics.NewProxyMetrics()
		prm.Counters = counters
		s.repo.Restore(prm)
	}
}
