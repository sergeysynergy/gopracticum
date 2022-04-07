package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

type Storage struct {
	gaugesMu sync.RWMutex
	gauges   map[string]metrics.Gauge

	countersMu sync.RWMutex
	counters   map[string]metrics.Counter
}

func New() *Storage {
	return &Storage{
		gauges:   make(map[string]metrics.Gauge, metrics.GaugeLen),
		counters: make(map[string]metrics.Counter, metrics.CounterLen),
	}
}

func NewWithGauges(gauges map[string]metrics.Gauge) *Storage {
	s := New()
	s.gauges = gauges
	return s
}

func NewWithCounters(counters map[string]metrics.Counter) *Storage {
	s := New()
	s.counters = counters
	return s
}

func (s *Storage) PutGauge(key string, val metrics.Gauge) {
	s.gaugesMu.Lock()
	defer s.gaugesMu.Unlock()

	s.gauges[key] = val
}

func (s *Storage) BulkPutGauge(gauges map[string]metrics.Gauge) {
	s.gaugesMu.Lock()
	defer s.gaugesMu.Unlock()

	s.gauges = gauges
}

func (s *Storage) GetGauge(key string) (metrics.Gauge, error) {
	s.gaugesMu.RLock()
	defer s.gaugesMu.RUnlock()

	gauge, ok := s.gauges[key]
	if !ok {
		return 0, fmt.Errorf("gauge metric with key '%s' not found", key)
	}

	return gauge, nil
}

func (s *Storage) Gauges() map[string]metrics.Gauge {
	s.gaugesMu.RLock()
	defer s.gaugesMu.RUnlock()

	return s.gauges
}

func (s *Storage) IncreaseCounter(key string) {
	s.countersMu.Lock()
	defer s.countersMu.Unlock()

	_, ok := s.counters[key]
	if !ok {
		s.counters[key] = 1
		return
	}

	s.counters[key] += 1
}

func (s *Storage) PostCounter(key string, val metrics.Counter) {
	s.countersMu.Lock()
	defer s.countersMu.Unlock()

	_, ok := s.counters[key]
	if !ok {
		s.counters[key] = val
		return
	}

	s.counters[key] += val
}

func (s *Storage) GetCounter(key string) (metrics.Counter, error) {
	s.countersMu.RLock()
	defer s.countersMu.RUnlock()

	counter, ok := s.counters[key]
	if !ok {
		return 0, fmt.Errorf("counter metric with key '%s' not found", key)
	}

	return counter, nil
}

func (s *Storage) Counters() map[string]metrics.Counter {
	s.countersMu.RLock()
	defer s.countersMu.RUnlock()

	return s.counters
}

// ToJSON Вывод содержимого хранилища в формате JSON для тестовых целей.
func (s *Storage) ToJSON() []byte {
	var b bytes.Buffer

	b.WriteString(`{"gauges":`)
	g, _ := json.Marshal(s.gauges)
	b.Write(g)
	b.WriteString(`},`)

	b.WriteString(`{"counters":`)
	c, _ := json.Marshal(s.counters)
	b.Write(c)
	b.WriteString(`}`)

	return b.Bytes()
}
