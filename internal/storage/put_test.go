package storage

import (
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
)

func ExampleStorage_Put() {
	st := New()
	alloc := metrics.Gauge(42.24)

	err := st.Put(metrics.Alloc, alloc)
	if err != nil {
		log.Fatalln("Failed to put metric: %w", err)
	}
}
