package agent

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sergeysynergy/gopracticum/internal/storage"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type Agent struct {
	client         *resty.Client
	storage        storage.Storer
	pollInterval   time.Duration
	reportInterval time.Duration
	protocol       string
	addr           string
}

type AgtOption func(agent *Agent)

func New(opts ...AgtOption) *Agent {
	rand.Seed(time.Now().UTC().UnixNano())

	const (
		defaultPollInterval   = 2 * time.Second  // частота обновления метрик из пакета `runtime`
		defaultReportInterval = 10 * time.Second // частота отправки метрик на сервер
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

func WithAddress(addr string) AgtOption {
	return func(a *Agent) {
		if addr != "" {
			a.addr = addr
		}
	}
}

func WithPollInterval(duration time.Duration) AgtOption {
	return func(a *Agent) {
		if duration > 0 {
			a.pollInterval = duration
		}
	}
}

func WithReportInterval(duration time.Duration) AgtOption {
	return func(a *Agent) {
		if duration > 0 {
			a.reportInterval = duration
		}
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
			a.Update()
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
			a.sendReport(ctx)
		case <-ctx.Done():
			log.Println("Штатное завершение работы отправки метрик")
			ticker.Stop()
			return
		}
	}
}

// Выполняем отправку запросов метрик на сервер.
func (a *Agent) sendReport(ctx context.Context) {
	for k, v := range a.storage.GetGauges() {
		gauge := float64(v)
		m := &metrics.Metrics{
			ID:    k,
			MType: "gauge",
			Value: &gauge,
		}
		err := a.sendJSONRequest(ctx, m)
		if err != nil {
			a.handleError(err)
			return
		}

	}
	for k, v := range a.storage.GetCounters() {
		counter := int64(v)
		m := &metrics.Metrics{
			ID:    k,
			MType: "counter",
			Delta: &counter,
		}
		err := a.sendJSONRequest(ctx, m)
		if err != nil {
			a.handleError(err)
			return
		}
	}

	log.Println("Выполнена отправка отчёта")
}

func (a *Agent) sendBasicRequest(ctx context.Context, wg *sync.WaitGroup, key string, value interface{}) {
	defer wg.Done()

	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	var endpoint string

	switch metric := value.(type) {
	case metrics.Gauge:
		endpoint = fmt.Sprintf("%s%s/update/%s/%s/%f", a.protocol, a.addr, "gauge", key, metric)
	case metrics.Counter:
		endpoint = fmt.Sprintf("%s%s/update/%s/%s/%d", a.protocol, a.addr, "counter", key, metric)
	default:
		a.handleError(fmt.Errorf("неизвестный тип метрики"))
		return
	}

	resp, err := a.client.R().
		SetContext(ctx).
		Post(endpoint)

	if err != nil {
		a.handleError(err)
		return
	}

	if resp.StatusCode() != http.StatusOK {
		respErr := fmt.Errorf("%v", resp.StatusCode())
		a.handleError(respErr)
		return
	}
}

func (a *Agent) sendJSONRequest(ctx context.Context, m *metrics.Metrics) error {
	endpoint := a.protocol + a.addr + "/update/"

	resp, err := a.client.R().
		SetContext(ctx).
		SetBody(m).
		Post(endpoint)

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("%v", resp.StatusCode())
	}

	return nil
}

func (a *Agent) handleError(err error) {
	log.Println("Ошибка -", err)
}

func (a *Agent) Update() {
	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	gauges := make(map[string]metrics.Gauge, metrics.GaugeLen)

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

	a.storage.BulkPutGauges(gauges)

	a.storage.IncreaseCounter(metrics.PollCount)

	log.Println("Выполнено обновление метрик")
}
