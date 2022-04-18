package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
)

const (
	GaugeLen   = 28
	CounterLen = 1

	Alloc         = "Alloc"
	BuckHashSys   = "BuckHashSys"
	Frees         = "Frees"
	GCCPUFraction = "GCCPUFraction"
	GCSys         = "GCSys"
	HeapAlloc     = "HeapAlloc"
	HeapIdle      = "HeapIdle"
	HeapInuse     = "HeapInuse"
	HeapObjects   = "HeapObjects"
	HeapReleased  = "HeapReleased"
	HeapSys       = "HeapSys"
	LastGC        = "LastGC"
	Lookups       = "Lookups"
	MCacheInuse   = "MCacheInuse"
	MCacheSys     = "MCacheSys"
	MSpanInuse    = "MSpanInuse"
	MSpanSys      = "MSpanSys"
	Mallocs       = "Mallocs"
	NextGC        = "NextGC"
	NumForcedGC   = "NumForcedGC"
	NumGC         = "NumGC"
	OtherSys      = "OtherSys"
	PauseTotalNs  = "PauseTotalNs"
	StackInuse    = "StackInuse"
	StackSys      = "StackSys"
	Sys           = "Sys"
	TotalAlloc    = "TotalAlloc"
	RandomValue   = "RandomValue"
	PollCount     = "PollCount"
)

type Gauge float64

func (g *Gauge) FromString(str string) error {
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}

	gauge := Gauge(val)
	*g = gauge

	return nil
}

type Counter int64

func (c *Counter) FromString(str string) error {
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return err
	}

	counter := Counter(val)
	*c = counter

	return nil
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

/*
func (m *Metrics) GaugeHash(key string) []byte {
	//gaugeKey := hash(fmt.Sprintf("%s:gauge:%f", id, value), key)

	msg := fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value)
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msg))

	return h.Sum(nil)
}

func (m *Metrics) CounterHash(key string) []byte {
	//counterKey := hash(fmt.Sprintf("%s:counter:%d", id, delta), key)

	msg := fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta)
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msg))

	return h.Sum(nil)
}

*/

type ProxyMetric struct {
	Gauges   map[string]Gauge
	Counters map[string]Counter
}

func GetGaugeHash(key, id string, val float64) string {
	//gaugeKey := hash(fmt.Sprintf("%s:gauge:%f", id, value), key)

	msg := fmt.Sprintf("%s:gauge:%f", id, val)
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msg))
	// переводим в 16-тиричный вид, чтобы хэш не пострадал при передаче в строковом представлении
	return hex.EncodeToString(h.Sum(nil))
}

func GetCounterHash(key, id string, val int64) string {
	//counterKey := hash(fmt.Sprintf("%s:counter:%d", id, delta), key)

	msg := fmt.Sprintf("%s:counter:%d", id, val)
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msg))
	// переводим в 16-тиричный вид, чтобы хэш не пострадал при передаче в строковом представлении
	return hex.EncodeToString(h.Sum(nil))
}
