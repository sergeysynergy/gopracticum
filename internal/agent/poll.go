package agent

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

// Выполняем регулярное обновление метрик пока не пришёл сигнал отмены.
func (a *Agent) pollTicker(ctx context.Context) {
	ticker := time.NewTicker(a.pollInterval)
	for {
		select {
		case <-ticker.C:
			a.pollUpdate()
		case <-ctx.Done():
			log.Println("Штатное завершение работы обновления метрик")
			ticker.Stop()
			return
		}
	}
}

func (a *Agent) pollUpdate() {
	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	prm := metrics.NewProxyMetrics()
	gauges := make(map[string]metrics.Gauge, metrics.TypeGaugeLen)

	gauges[metrics.Alloc] = metrics.Gauge(ms.Alloc)
	gauges[metrics.BuckHashSys] = metrics.Gauge(ms.BuckHashSys)
	gauges[metrics.Frees] = metrics.Gauge(ms.Frees)
	gauges[metrics.GCCPUFraction] = metrics.Gauge(ms.GCCPUFraction)
	gauges[metrics.GCSys] = metrics.Gauge(ms.GCSys)
	gauges[metrics.HeapAlloc] = metrics.Gauge(ms.HeapAlloc)
	gauges[metrics.HeapIdle] = metrics.Gauge(ms.HeapIdle)
	gauges[metrics.HeapInuse] = metrics.Gauge(ms.HeapInuse)
	gauges[metrics.HeapObjects] = metrics.Gauge(ms.HeapObjects)
	gauges[metrics.HeapReleased] = metrics.Gauge(ms.HeapReleased)
	gauges[metrics.HeapSys] = metrics.Gauge(ms.HeapSys)
	gauges[metrics.LastGC] = metrics.Gauge(ms.LastGC)
	gauges[metrics.Lookups] = metrics.Gauge(ms.Lookups)
	gauges[metrics.MCacheInuse] = metrics.Gauge(ms.MCacheInuse)
	gauges[metrics.MCacheSys] = metrics.Gauge(ms.MCacheSys)
	gauges[metrics.MSpanInuse] = metrics.Gauge(ms.MSpanInuse)
	gauges[metrics.MSpanSys] = metrics.Gauge(ms.MSpanSys)
	gauges[metrics.Mallocs] = metrics.Gauge(ms.Mallocs)
	gauges[metrics.NextGC] = metrics.Gauge(ms.NextGC)
	gauges[metrics.NumForcedGC] = metrics.Gauge(ms.NumForcedGC)
	gauges[metrics.NumGC] = metrics.Gauge(ms.NumGC)
	gauges[metrics.OtherSys] = metrics.Gauge(ms.OtherSys)
	gauges[metrics.PauseTotalNs] = metrics.Gauge(ms.PauseTotalNs)
	gauges[metrics.StackInuse] = metrics.Gauge(ms.StackInuse)
	gauges[metrics.StackSys] = metrics.Gauge(ms.StackSys)
	gauges[metrics.Sys] = metrics.Gauge(ms.Sys)
	gauges[metrics.TotalAlloc] = metrics.Gauge(ms.TotalAlloc)
	gauges[metrics.RandomValue] = metrics.Gauge(rand.Float64())

	prm.Gauges = gauges

	prm.Counters[metrics.PollCount] = 1

	err := a.storage.PutMetrics(prm)
	if err != nil {
		a.handleError(fmt.Errorf("ошибка обновления метрик - %w", err))
	}

	log.Println("Выполнено обновление метрик")
}
