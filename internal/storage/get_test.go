package storage

import (
	"fmt"
	"log"

	"github.com/sergeysynergy/metricser/internal/data/repository/memory"
	"github.com/sergeysynergy/metricser/pkg/metrics"
)

func ExampleStorage_Get() {
	st := New(memory.New(), nil)

	alloc, err := st.Get(metrics.Alloc)
	if err != nil {
		log.Fatalln("Failed to get metric: %w", err)
	}

	fmt.Println(alloc)
}
