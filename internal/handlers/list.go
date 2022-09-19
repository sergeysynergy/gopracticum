package handlers

import (
	"bytes"
	"fmt"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"net/http"
	"sort"
	"strconv"
)

// List Возвращает список со значением всех метрик.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", textHTML)
	w.WriteHeader(http.StatusOK)

	var b bytes.Buffer
	b.WriteString("<h1>Current metrics data:</h1>")

	type gauge struct {
		key   string
		value float64
	}

	mcs, err := h.uc.GetMetrics()
	if err != nil {
		h.errorJSON(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	gauges := make([]gauge, 0, metrics.TypeGaugeLen)
	for k, val := range mcs.Gauges {
		gauges = append(gauges, gauge{key: k, value: float64(val)})
	}
	sort.Slice(gauges, func(i, j int) bool { return gauges[i].key < gauges[j].key })

	b.WriteString(`<div><h2>Gauges</h2>`)
	for _, g := range gauges {
		val := strconv.FormatFloat(g.value, 'f', -1, 64)
		b.WriteString(fmt.Sprintf("<div>%s - %v</div>", g.key, val))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<div><h2>Counters</h2>`)
	for k, val := range mcs.Counters {
		b.WriteString(fmt.Sprintf("<div>%s - %d</div>", k, val))
	}
	b.WriteString(`</div>`)

	w.Write(b.Bytes())
}
