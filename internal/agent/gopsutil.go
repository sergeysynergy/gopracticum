package agent

import (
	"context"
	"fmt"
	"github.com/shirou/gopsutil/v3/mem"
	"log"
	"math/rand"
	"time"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

// Выполняем регулярное обновление метрик посредством пакета `gopsutil` пока не пришёл сигнал отмены.
func (a *Agent) gopsutilTicker(ctx context.Context) {
	ticker := time.NewTicker(a.pollInterval)
	for {
		select {
		case <-ticker.C:
			a.gopsutilUpdate()
		case <-ctx.Done():
			log.Println("Штатное завершение работы обновления метрик")
			ticker.Stop()
			return
		}
	}
}

func (a *Agent) gopsutilUpdate() {
	prm := metrics.NewProxyMetrics()
	gauges := make(map[string]metrics.Gauge, 3)

	//c, err := cpu.Counts(true)
	//if err != nil {
	//	a.handleError(fmt.Errorf("ошибка получения метрик посредством пакета `gopsutil` - %w", err))
	//}
	c := rand.Intn(12)
	gauges[metrics.CPUutilization1] = metrics.Gauge(c)

	v, err := mem.VirtualMemory()
	if err != nil {
		a.handleError(fmt.Errorf("ошибка получения метрик посредством пакета `gopsutil` - %w", err))
	}
	gauges[metrics.TotalMemory] = metrics.Gauge(v.Total)
	gauges[metrics.FreeMemory] = metrics.Gauge(v.Free)

	prm.Gauges = gauges

	err = a.storage.PutMetrics(prm)
	if err != nil {
		a.handleError(fmt.Errorf("ошибка обновления метрик посредством пакета `gopsutil` - %w", err))
	}

	log.Println("Выполнено обновление метрик посредством пакета `gopsutil`")
}
