package handlers

import (
	"html/template"
	"net/http"
	"sort"
)

const listTemplate = `<h1>Current metrics data</h1>
{{if .Gauges}}
<h2>Gauges:</h1>
{{range .Gauges}}<div>{{.Key}} - {{.Value}}</div>{{end}}
{{end}}
{{if .Counters}}
<h2>Counters:</h1>{{range .Counters}}<div>{{.Key}} - {{.Delta}}</div>{{end}}
{{end}}
`

// List Возвращает список со значением всех метрик.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	type (
		gauge struct {
			Key   string
			Value float64
		}
		counter struct {
			Key   string
			Delta int64
		}
		metrics struct {
			Gauges   []gauge
			Counters []counter
		}
	)

	w.Header().Set("content-type", textHTML)
	w.WriteHeader(http.StatusOK)

	prm, err := h.uc.GetMetrics()
	if err != nil {
		h.errorJSON(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	gauges := make([]gauge, 0, len(prm.Gauges))
	for k, val := range prm.Gauges {
		gauges = append(gauges, gauge{Key: k, Value: float64(val)})
	}
	sort.Slice(gauges, func(i, j int) bool { return gauges[i].Key < gauges[j].Key })

	counters := make([]counter, 0, len(prm.Counters))
	for k, val := range prm.Counters {
		counters = append(counters, counter{Key: k, Delta: int64(val)})
	}
	sort.Slice(counters, func(i, j int) bool { return counters[i].Key < counters[j].Key })

	mcs := metrics{
		Gauges:   gauges,
		Counters: counters,
	}

	t, err := template.New("list").Parse(listTemplate)
	if err != nil {
		panic(err)
	}
	err = t.ExecuteTemplate(w, "list", mcs)
	if err != nil {
		panic(err)
	}
}
