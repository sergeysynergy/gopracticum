package main

import (
	"context"
	"fmt"
	"io"
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

const (
	gaugeLen   = 28
	counterLen = 1
)

type gauge float64

type counter int64

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

type AgentConfig struct {
	pollInterval   time.Duration // частота обновления метрик из пакета `runtime`
	reportInterval time.Duration // частота отправки метрик на сервер
	address        string        // адрес сервера куда отправлять метрики
	port           string        // порт сервера
}

type Agent struct {
	cfg     AgentConfig
	metrics Metrics
	client  http.Client
}

// Выполняем регулярное обновление метрик пока не пришёл сигнал отмены.
func (a *Agent) pollHandler(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.pollInterval)
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
	ticker := time.NewTicker(a.cfg.reportInterval)
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
		endpoint = fmt.Sprintf("http://%s%s/update/%s/%s/%f", a.cfg.address, a.cfg.port, "gauge", key, metric)
	case counter:
		endpoint = fmt.Sprintf("http://%s%s/update/%s/%s/%d", a.cfg.address, a.cfg.port, "counter", key, metric)
	default:
		handleError(fmt.Errorf("неизвестный тип метрики"))
		return
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	req.Header.Set("Content-Type", "text/plain")

	response, err := a.client.Do(req)
	if err != nil {
		handleError(err)
		return
	}

	// Чтобы повторно использовать кешированное TCP-соединение, клиент должен обязательно прочитать тело ответа
	// до конца и закрыть, даже если оно не нужно.
	_, err = io.Copy(io.Discard, response.Body)
	if err != nil {
		handleError(err)
		return
	}
	defer response.Body.Close()
}

func handleError(err error) {
	log.Println("Ошибка -", err)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	cfg := AgentConfig{
		pollInterval:   2 * time.Second,
		reportInterval: 10 * time.Second,
		address:        "127.0.0.1",
		port:           ":8080",
	}
	agent := Agent{
		cfg: cfg,
		metrics: Metrics{
			gauges:   make(map[string]gauge, gaugeLen),
			counters: make(map[string]counter, counterLen),
		},
		client: http.Client{
			Timeout: 4 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: 30,
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Функцию cancel нужно обязательно выполнить в коде, иначе сборщик мусора не удалит созданный дочерний контекст
	// и произойдёт утечка памяти.
	defer cancel()

	go agent.pollHandler(ctx)
	go agent.reportHandler(ctx)

	// Агент должен штатно завершаться по сигналам: syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	s := <-c
	log.Println("Получен сигнал завершения работы:", s)
	log.Println("Работа агента штатно завершена")
}
