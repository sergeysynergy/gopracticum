package storage

import (
	"github.com/sergeysynergy/metricser/internal/data/repository/memory"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
)

func ExampleStorage_Put() {
	st := New(memory.New(), nil)
	alloc := metrics.Gauge(42.24)

	err := st.Put(metrics.Alloc, alloc)
	if err != nil {
		log.Fatalln("Failed to put metric: %w", err)
	}
}
