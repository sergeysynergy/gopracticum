package agent

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/sergeysynergy/gopracticum/internal/storage"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

type Agent struct {
	client         *resty.Client
	storage        storage.Storer
	pollInterval   time.Duration
	reportInterval time.Duration
	protocol       string
	addr           string
	key            string
}

type Option func(agent *Agent)

func New(opts ...Option) *Agent {
	rand.Seed(time.Now().UTC().UnixNano())

	const (
		defaultReportInterval = 10 * time.Second // частота отправки метрик на сервер
		defaultPollInterval   = 2 * time.Second  // частота обновления метрик из пакета `runtime`
		defaultAddress        = "127.0.0.1:8080"
		defaultProtocol       = "http://"
		defaultTimeout        = 4 * time.Second
	)

	a := &Agent{
		client:         resty.New(),
		storage:        storage.New(),
		pollInterval:   defaultPollInterval,
		reportInterval: defaultReportInterval,
		protocol:       defaultProtocol,
		addr:           defaultAddress,
	}
	a.client.SetTimeout(defaultTimeout)

	// Применяем в цикле каждую опцию
	for _, opt := range opts {
		// вызываем функцию, предоставляющую экземпляр *Agent в качестве аргумента
		opt(a)
	}

	// вернуть измененный экземпляр Server
	return a
}

func WithAddress(addr string) Option {
	return func(a *Agent) {
		if addr != "" {
			a.addr = addr
		}
	}
}

func WithPollInterval(duration time.Duration) Option {
	return func(a *Agent) {
		if duration > 0 {
			a.pollInterval = duration
		}
	}
}

func WithReportInterval(duration time.Duration) Option {
	return func(a *Agent) {
		if duration > 0 {
			a.reportInterval = duration
		}
	}
}

func WithKey(key string) Option {
	return func(a *Agent) {
		a.key = key
	}
}

func (a *Agent) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	// Функцию cancel нужно обязательно выполнить в коде, иначе сборщик мусора не удалит созданный дочерний контекст
	// и произойдёт утечка памяти.
	defer cancel()

	go a.pollHandler(ctx)
	go a.reportHandler(ctx)

	// Агент должен штатно завершаться по сигналам: syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-c
	log.Println("Получен сигнал завершения работы:", sig)
	log.Println("Работа агента штатно завершена")
}

// Выполняем регулярное обновление метрик пока не пришёл сигнал отмены.
func (a *Agent) pollHandler(ctx context.Context) {
	ticker := time.NewTicker(a.pollInterval)
	for {
		select {
		case <-ticker.C:
			a.Update(ctx)
		case <-ctx.Done():
			log.Println("Штатное завершение работы обновления метрик")
			ticker.Stop()
			return
		}
	}
}

// Выполняем регулярную отправку метрик на сервер пока не пришёл сигнал отмены.
func (a *Agent) reportHandler(ctx context.Context) {
	ticker := time.NewTicker(a.reportInterval)
	for {
		select {
		case <-ticker.C:
			a.sendReportUpdates(ctx)
		case <-ctx.Done():
			log.Println("Штатное завершение работы отправки метрик")
			ticker.Stop()
			return
		}
	}
}

// Выполняем отправку запросов метрик на сервер.
func (a *Agent) sendReportUpdates(ctx context.Context) {
	hm := make([]metrics.Metrics, 0, metrics.TypeGaugeLen+metrics.TypeCounterLen)

	prm, err := a.storage.GetMetrics(ctx)
	if err != nil {
		a.handleError(err)
		return
	}

	var hash string

	for k, v := range prm.Gauges {
		value := float64(v)

		// добавляем хэш, если задан ключ key
		if a.key != "" {
			hash = metrics.GaugeHash(a.key, k, value)
		}

		hm = append(hm, metrics.Metrics{
			ID:    k,
			MType: metrics.TypeGauge,
			Value: &value,
			Hash:  hash,
		})
	}

	for k, v := range prm.Counters {
		delta := int64(v)

		// добавляем хэш, если задан ключ key
		if a.key != "" {
			hash = metrics.CounterHash(a.key, k, delta)
		}

		hm = append(hm, metrics.Metrics{
			ID:    k,
			MType: metrics.TypeCounter,
			Delta: &delta,
			Hash:  hash,
		})
	}

	if len(hm) == 0 {
		log.Println("[WARNING] Пустой массив метрик, отправлять нечего")
		return
	}

	_, err = a.sendUpdates(ctx, hm)
	if err != nil {
		a.handleError(err)
		return
	}

	log.Println("Выполнена отправка отчёта")
}

func (a *Agent) handleError(err error) {
	log.Println("Ошибка -", err)
}

func (a *Agent) Update(ctx context.Context) {
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

	err := a.storage.PutMetrics(ctx, prm)
	if err != nil {
		a.handleError(err)
	}

	log.Println("Выполнено обновление метрик")
}

func (a *Agent) sendUpdates(ctx context.Context, hm []metrics.Metrics) (*resty.Response, error) {
	endpoint := a.protocol + a.addr + "/updates/"

	resp, err := a.client.R().
		SetHeader("Accept", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetContext(ctx).
		SetBody(hm).
		Post(endpoint)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return resp, fmt.Errorf("invalid status code %v", resp.StatusCode())
	}

	return resp, nil
}
