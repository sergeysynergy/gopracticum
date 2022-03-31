package metrics

import (
	"strconv"
	"sync"
)

const (
	GaugeLen   = 28
	CounterLen = 1

	Alloc         = Name("Alloc")
	BuckHashSys   = Name("BuckHashSys")
	Frees         = Name("Frees")
	GCCPUFraction = Name("GCCPUFraction")
	GCSys         = Name("GCSys")
	HeapAlloc     = Name("HeapAlloc")
	HeapIdle      = Name("HeapIdle")
	HeapInuse     = Name("HeapInuse")
	HeapObjects   = Name("HeapObjects")
	HeapReleased  = Name("HeapReleased")
	HeapSys       = Name("HeapSys")
	LastGC        = Name("LastGC")
	Lookups       = Name("Lookups")
	MCacheInuse   = Name("MCacheInuse")
	MCacheSys     = Name("MCacheSys")
	MSpanInuse    = Name("MSpanInuse")
	MSpanSys      = Name("MSpanSys")
	Mallocs       = Name("Mallocs")
	NextGC        = Name("NextGC")
	NumForcedGC   = Name("NumForcedGC")
	NumGC         = Name("NumGC")
	OtherSys      = Name("OtherSys")
	PauseTotalNs  = Name("PauseTotalNs")
	StackInuse    = Name("StackInuse")
	StackSys      = Name("StackSys")
	Sys           = Name("Sys")
	TotalAlloc    = Name("TotalAlloc")
	RandomValue   = Name("RandomValue")
	PollCount     = Name("PollCount")
)

type Name string

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
	sync.RWMutex
	Gauges   map[Name]Gauge
	Counters map[Name]Counter
}

func New() *Metrics {
	return &Metrics{
		Gauges:   make(map[Name]Gauge, GaugeLen),
		Counters: make(map[Name]Counter, CounterLen),
	}
}
