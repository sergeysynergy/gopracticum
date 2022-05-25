package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
)

const (
	TypeGauge      = "gauge"
	TypeCounter    = "counter"
	TypeGaugeLen   = 31
	TypeCounterLen = 1

	Alloc           = "Alloc"
	BuckHashSys     = "BuckHashSys"
	Frees           = "Frees"
	GCCPUFraction   = "GCCPUFraction"
	GCSys           = "GCSys"
	HeapAlloc       = "HeapAlloc"
	HeapIdle        = "HeapIdle"
	HeapInuse       = "HeapInuse"
	HeapObjects     = "HeapObjects"
	HeapReleased    = "HeapReleased"
	HeapSys         = "HeapSys"
	LastGC          = "LastGC"
	Lookups         = "Lookups"
	MCacheInuse     = "MCacheInuse"
	MCacheSys       = "MCacheSys"
	MSpanInuse      = "MSpanInuse"
	MSpanSys        = "MSpanSys"
	Mallocs         = "Mallocs"
	NextGC          = "NextGC"
	NumForcedGC     = "NumForcedGC"
	NumGC           = "NumGC"
	OtherSys        = "OtherSys"
	PauseTotalNs    = "PauseTotalNs"
	StackInuse      = "StackInuse"
	StackSys        = "StackSys"
	Sys             = "Sys"
	TotalAlloc      = "TotalAlloc"
	RandomValue     = "RandomValue"
	PollCount       = "PollCount"
	TotalMemory     = "TotalMemory"
	FreeMemory      = "FreeMemory"
	CPUutilization1 = "CPUutilization1"
)

var Gauges = map[string]bool{
	Alloc:           true,
	BuckHashSys:     true,
	Frees:           true,
	GCCPUFraction:   true,
	GCSys:           true,
	HeapAlloc:       true,
	HeapIdle:        true,
	HeapInuse:       true,
	HeapObjects:     true,
	HeapReleased:    true,
	HeapSys:         true,
	LastGC:          true,
	Lookups:         true,
	MCacheInuse:     true,
	MCacheSys:       true,
	MSpanInuse:      true,
	MSpanSys:        true,
	Mallocs:         true,
	NextGC:          true,
	NumForcedGC:     true,
	NumGC:           true,
	OtherSys:        true,
	PauseTotalNs:    true,
	StackInuse:      true,
	StackSys:        true,
	Sys:             true,
	TotalAlloc:      true,
	RandomValue:     true,
	TotalMemory:     true,
	FreeMemory:      true,
	CPUutilization1: true,
}

var Counters = map[string]bool{
	PollCount: true,
}

func IsKnown(id string) bool {
	if _, ok := Gauges[id]; ok {
		return true
	}

	if _, ok := Counters[id]; ok {
		return true
	}

	return false
}

var (
	ErrNotImplemented = errors.New("metrics: metric not implemented")
	ErrNotFound       = errors.New("metrics: metric not found")
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

type ProxyMetrics struct {
	Gauges   map[string]Gauge
	Counters map[string]Counter
}

func NewProxyMetrics() ProxyMetrics {
	return ProxyMetrics{
		Gauges:   make(map[string]Gauge, TypeGaugeLen),
		Counters: make(map[string]Counter, TypeCounterLen),
	}
}

func GaugeHash(key, id string, value float64) string {
	msg := fmt.Sprintf("%s:gauge:%f", id, value)
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msg))
	// переводим в 16-тиричный вид, чтобы хэш не пострадал при передаче в строковом представлении
	return hex.EncodeToString(h.Sum(nil))
}

func CounterHash(key, id string, delta int64) string {
	msg := fmt.Sprintf("%s:counter:%d", id, delta)
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msg))
	// переводим в 16-тиричный вид, чтобы хэш не пострадал при передаче в строковом представлении
	return hex.EncodeToString(h.Sum(nil))
}
