package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sergeysynergy/gopracticum/internal/storage"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

type Gauge struct {
	*storage.Storage
}

type Counter struct {
	*storage.Storage
}
type Check struct {
	*storage.Storage
}

func (g *Gauge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path, "/")

	if len(params) != 5 {
		http.Error(w, "Wrong parameters number!", http.StatusNotFound)
		return
	}
	if params[4] == "" {
		http.Error(w, "Wrong value!", http.StatusNotAcceptable)
		return
	}

	metricName := metrics.Name(params[3])

	var gauge metrics.Gauge
	err := gauge.FromString(params[4])
	if err != nil {
		msg := fmt.Sprintf("value %v not acceptable - %v", params[4], err)
		http.Error(w, msg, http.StatusNotAcceptable)
		return
	}
	g.Put(metricName, gauge)

	w.WriteHeader(http.StatusOK)
}

func (c *Counter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path, "/")

	if len(params) != 5 {
		http.Error(w, "Wrong parameters number!", http.StatusNotFound)
		return
	}
	if params[4] == "" {
		http.Error(w, "Wrong value!", http.StatusNotAcceptable)
		return
	}

	metricName := metrics.Name(params[3])

	var counter metrics.Counter
	err := counter.FromString(params[4])
	if err != nil {
		msg := fmt.Sprintf("value %v not acceptable - %v", params[4], err)
		http.Error(w, msg, http.StatusNotAcceptable)
		return
	}
	c.Count(metricName, counter)

	w.WriteHeader(http.StatusOK)
}

// Хэндлер для визуальной проверки результатов работы.
func (c *Check) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(c.ToJSON())
}

func NotImplemented(w http.ResponseWriter, r *http.Request) {
	err := fmt.Errorf("not implemented")
	http.Error(w, err.Error(), http.StatusNotImplemented)
}
