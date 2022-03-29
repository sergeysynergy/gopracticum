package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

const (
	gaugeLen   = 28
	counterLen = 1
)

type gauge float64

type counter int64

type Config struct {
	PollInterval   time.Duration // частота обновления метрик из пакета `runtime`
	ReportInterval time.Duration // частота отправки метрик на сервер
	Address        string        // адрес сервера куда отправлять метрики
	Port           string        // порт сервера
}

type Agent struct {
	Cfg     Config
	metrics Metrics
	client  http.Client
}

func New(cfg Config) (*Agent, error) {
	if cfg.PollInterval == 0 {
		return nil, fmt.Errorf("необходимо задать PollInterval")
	}
	if cfg.ReportInterval == 0 {
		return nil, fmt.Errorf("необходимо задать ReportInterval")
	}
	if cfg.Address == "" {
		return nil, fmt.Errorf("необходимо задать адрес сервера")
	}
	if cfg.Port == "" {
		return nil, fmt.Errorf("необходимо задать порт сервера")
	}

	a := &Agent{
		Cfg: cfg,
		metrics: Metrics{
			gauges:   make(map[string]gauge, gaugeLen),
			counters: make(map[string]counter, counterLen),
		},
		client: http.Client{
			Timeout: 4 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: gaugeLen + counterLen,
			},
		},
	}

	return a, nil
}

func (a *Agent) Run() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	go a.pollHandler(ctx)
	go a.reportHandler(ctx)

	return cancel
}

// Выполняем регулярное обновление метрик пока не пришёл сигнал отмены.
func (a *Agent) pollHandler(ctx context.Context) {
	ticker := time.NewTicker(a.Cfg.PollInterval)
	for {
		select {
		case <-ticker.C:
			a.metrics.Update()
		case <-ctx.Done():
			log.Println("Штатное завершение работы обновления метрик")
			ticker.Stop()
			return
		}
	}
}

// Выполняем регулярную отправку метрик на сервер пока не пришёл сигнал отмены.
func (a *Agent) reportHandler(ctx context.Context) {
	ticker := time.NewTicker(a.Cfg.ReportInterval)
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
	a.metrics.mu.RLock()
	defer a.metrics.mu.RUnlock()

	wg := &sync.WaitGroup{}
	for key, val := range a.metrics.gauges {
		wg.Add(1)
		go a.sendRequest(ctx, wg, key, val)
	}
	for key, val := range a.metrics.counters {
		wg.Add(1)
		go a.sendRequest(ctx, wg, key, val)
	}
	wg.Wait()

	log.Println("Выполнена отправка отчёта")
}

func (a *Agent) sendRequest(ctx context.Context, wg *sync.WaitGroup, key string, value interface{}) {
	defer wg.Done()

	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	var endpoint string

	switch metric := value.(type) {
	case gauge:
		endpoint = fmt.Sprintf("http://%s%s/update/%s/%s/%f", a.Cfg.Address, a.Cfg.Port, "gauge", key, metric)
	case counter:
		endpoint = fmt.Sprintf("http://%s%s/update/%s/%s/%d", a.Cfg.Address, a.Cfg.Port, "counter", key, metric)
	default:
		a.handleError(fmt.Errorf("неизвестный тип метрики"))
		return
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	req.Header.Set("Content-Type", "text/plain")

	response, err := a.client.Do(req)
	if err != nil {
		a.handleError(err)
		return
	}

	// Чтобы повторно использовать кешированное TCP-соединение, клиент должен обязательно прочитать тело ответа
	// до конца и закрыть, даже если оно не нужно.
	_, err = io.Copy(io.Discard, response.Body)
	if err != nil {
		a.handleError(err)
		return
	}
	defer response.Body.Close()
}

func (a *Agent) handleError(err error) {
	log.Println("Ошибка -", err)
}

type Metrics struct {
	mu       sync.RWMutex
	gauges   map[string]gauge
	counters map[string]counter
}

func (m *Metrics) Update() {
	m.mu.Lock()
	defer m.mu.Unlock()

	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	m.gauges["Alloc"] = gauge(ms.Alloc)
	m.gauges["BuckHashSys"] = gauge(ms.BuckHashSys)
	m.gauges["Frees"] = gauge(ms.Frees)
	m.gauges["GCCPUFraction"] = gauge(ms.GCCPUFraction)
	m.gauges["GCSys"] = gauge(ms.GCSys)
	m.gauges["HeapAlloc"] = gauge(ms.HeapAlloc)
	m.gauges["HeapIdle"] = gauge(ms.HeapIdle)
	m.gauges["HeapInuse"] = gauge(ms.HeapInuse)
	m.gauges["HeapObjects"] = gauge(ms.HeapObjects)
	m.gauges["HeapReleased"] = gauge(ms.HeapReleased)
	m.gauges["HeapSys"] = gauge(ms.HeapSys)
	m.gauges["LastGC"] = gauge(ms.LastGC)
	m.gauges["Lookups"] = gauge(ms.Lookups)
	m.gauges["MCacheInuse"] = gauge(ms.MCacheInuse)
	m.gauges["MCacheSys"] = gauge(ms.MCacheSys)
	m.gauges["MSpanInuse"] = gauge(ms.MSpanInuse)
	m.gauges["MSpanSys"] = gauge(ms.MSpanSys)
	m.gauges["Mallocs"] = gauge(ms.Mallocs)
	m.gauges["NextGC"] = gauge(ms.NextGC)
	m.gauges["NumForcedGC"] = gauge(ms.NumForcedGC)
	m.gauges["NumGC"] = gauge(ms.NumGC)
	m.gauges["OtherSys"] = gauge(ms.OtherSys)
	m.gauges["PauseTotalNs"] = gauge(ms.PauseTotalNs)
	m.gauges["StackInuse"] = gauge(ms.StackInuse)
	m.gauges["StackSys"] = gauge(ms.StackSys)
	m.gauges["Sys"] = gauge(ms.Sys)
	m.gauges["TotalAlloc"] = gauge(ms.TotalAlloc)
	m.gauges["RandomValue"] = gauge(rand.Float64())

	m.counters["PollCount"] += 1

	log.Println("Выполнено обновление метрик")
}
