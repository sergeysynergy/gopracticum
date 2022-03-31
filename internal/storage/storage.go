package storage

import (
	"bytes"
	"encoding/json"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

type Storage struct {
	*metrics.Metrics
}

func New() *Storage {
	return &Storage{
		Metrics: metrics.New(),
	}
}

func (s *Storage) Put(key metrics.Name, val metrics.Gauge) {
	s.Lock()
	defer s.Unlock()

	s.Gauges[key] = val
}

func (s *Storage) Count(key metrics.Name, val metrics.Counter) {
	s.Lock()
	defer s.Unlock()

	_, ok := s.Counters[key]
	if !ok {
		s.Counters[key] = val
		return
	}

	s.Counters[key] += val
}

// ToJSON Вывод содержимого хранилища в формате JSON для тестовых целей.
func (s *Storage) ToJSON() []byte {
	var b bytes.Buffer

	b.WriteString(`{"gauges":`)
	g, _ := json.Marshal(s.Gauges)
	b.Write(g)
	b.WriteString(`},`)

	b.WriteString(`{"counters":`)
	c, _ := json.Marshal(s.Counters)
	b.Write(c)
	b.WriteString(`}`)

	return b.Bytes()
}
