package storage

import (
	"fmt"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
)

func ExampleStorage_Restore() {
	st := New()

	prm := &metrics.ProxyMetrics{
		Gauges: map[string]metrics.Gauge{
			metrics.Alloc:       42.24,
			metrics.BuckHashSys: 77,
		},
		Counters: map[string]metrics.Counter{
			metrics.PollCount: 1,
		},
	}

	st.Restore(prm)

	prm, err := st.GetMetrics()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Metrics after restore: %#v", prm)
}
