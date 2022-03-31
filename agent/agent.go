package agent

import (
	"context"
	"fmt"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
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

type Config struct {
	PollInterval   time.Duration // частота обновления метрик из пакета `runtime`
	ReportInterval time.Duration // частота отправки метрик на сервер
	Address        string        // адрес сервера куда отправлять метрики
	Port           string        // порт сервера
}

type Agent struct {
	*metrics.Metrics
	Cfg    Config
	client http.Client
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
		Cfg:     cfg,
		Metrics: metrics.New(),
		client: http.Client{
			Timeout: 4 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: metrics.GaugeLen + metrics.CounterLen,
			},
		},
	}

	return a, nil
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
	ticker := time.NewTicker(a.Cfg.PollInterval)
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
	a.RLock()
	defer a.RUnlock()

	wg := &sync.WaitGroup{}
	for key, val := range a.Gauges {
		wg.Add(1)
		go a.sendRequest(ctx, wg, key, val)
	}
	for key, val := range a.Counters {
		wg.Add(1)
		go a.sendRequest(ctx, wg, key, val)
	}
	wg.Wait()

	log.Println("Выполнена отправка отчёта")
}

func (a *Agent) sendRequest(ctx context.Context, wg *sync.WaitGroup, key metrics.Name, value interface{}) {
	defer wg.Done()

	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	var endpoint string

	switch metric := value.(type) {
	case metrics.Gauge:
		endpoint = fmt.Sprintf("http://%s%s/update/%s/%s/%f", a.Cfg.Address, a.Cfg.Port, "gauge", key, metric)
	case metrics.Counter:
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

func (a *Agent) Update() {
	a.Lock()
	defer a.Unlock()

	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	a.Gauges[metrics.Alloc] = metrics.Gauge(ms.Alloc)
	a.Gauges[metrics.BuckHashSys] = metrics.Gauge(ms.BuckHashSys)
	a.Gauges[metrics.Frees] = metrics.Gauge(ms.Frees)
	a.Gauges[metrics.GCCPUFraction] = metrics.Gauge(ms.GCCPUFraction)
	a.Gauges[metrics.GCSys] = metrics.Gauge(ms.GCSys)
	a.Gauges[metrics.HeapAlloc] = metrics.Gauge(ms.HeapAlloc)
	a.Gauges[metrics.HeapIdle] = metrics.Gauge(ms.HeapIdle)
	a.Gauges[metrics.HeapInuse] = metrics.Gauge(ms.HeapInuse)
	a.Gauges[metrics.HeapObjects] = metrics.Gauge(ms.HeapObjects)
	a.Gauges[metrics.HeapReleased] = metrics.Gauge(ms.HeapReleased)
	a.Gauges[metrics.HeapSys] = metrics.Gauge(ms.HeapSys)
	a.Gauges[metrics.LastGC] = metrics.Gauge(ms.LastGC)
	a.Gauges[metrics.Lookups] = metrics.Gauge(ms.Lookups)
	a.Gauges[metrics.MCacheInuse] = metrics.Gauge(ms.MCacheInuse)
	a.Gauges[metrics.MCacheSys] = metrics.Gauge(ms.MCacheSys)
	a.Gauges[metrics.MSpanInuse] = metrics.Gauge(ms.MSpanInuse)
	a.Gauges[metrics.MSpanSys] = metrics.Gauge(ms.MSpanSys)
	a.Gauges[metrics.Mallocs] = metrics.Gauge(ms.Mallocs)
	a.Gauges[metrics.NextGC] = metrics.Gauge(ms.NextGC)
	a.Gauges[metrics.NumForcedGC] = metrics.Gauge(ms.NumForcedGC)
	a.Gauges[metrics.NumGC] = metrics.Gauge(ms.NumGC)
	a.Gauges[metrics.OtherSys] = metrics.Gauge(ms.OtherSys)
	a.Gauges[metrics.PauseTotalNs] = metrics.Gauge(ms.PauseTotalNs)
	a.Gauges[metrics.StackInuse] = metrics.Gauge(ms.StackInuse)
	a.Gauges[metrics.StackSys] = metrics.Gauge(ms.StackSys)
	a.Gauges[metrics.Sys] = metrics.Gauge(ms.Sys)
	a.Gauges[metrics.TotalAlloc] = metrics.Gauge(ms.TotalAlloc)
	a.Gauges[metrics.RandomValue] = metrics.Gauge(rand.Float64())

	a.Counters[metrics.PollCount] += 1

	log.Println("Выполнено обновление метрик")
}
