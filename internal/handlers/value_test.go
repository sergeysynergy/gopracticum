package handlers

import (
	"github.com/go-resty/resty/v2"
	"github.com/sergeysynergy/gopracticum/internal/filestore"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sergeysynergy/gopracticum/internal/storage"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

func TestValueContentType(t *testing.T) {
	h := New()
	ts := httptest.NewServer(h.router)
	defer ts.Close()

	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept", applicationJSON).
		SetHeader("Content-Type", "bad type").
		Post(ts.URL + "/value/")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode())
	assert.Equal(t, applicationJSON, resp.Header().Get("Content-Type"))
}

func TestValueUnmarshalError(t *testing.T) {
	h := New()
	ts := httptest.NewServer(h.router)
	defer ts.Close()

	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept", applicationJSON).
		SetHeader("Content-Type", applicationJSON).
		SetBody("{bad bad json").
		Post(ts.URL + "/value/")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotAcceptable, resp.StatusCode())
	assert.Equal(t, applicationJSON, resp.Header().Get("Content-Type"))
}

func TestValue(t *testing.T) {
	key := "Passw0rd33"
	type myMetrics struct {
		ID    string  `json:"id"`
		MType string  `json:"type"`
		Delta int64   `json:"delta,omitempty"`
		Value float64 `json:"value,omitempty"`
	}
	type want struct {
		statusCode int
		body       myMetrics
	}
	tests := []struct {
		name    string
		handler *Handler
		body    metrics.Metrics
		key     string
		want    want
	}{
		{
			name:    "Metric type needed",
			handler: New(),
			body:    metrics.Metrics{MType: ""},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "Not implemented",
			handler: New(),
			body:    metrics.Metrics{MType: "unknown"},
			want: want{
				statusCode: http.StatusNotImplemented,
			},
		},
		{
			name:    "Gauge not found",
			handler: New(),
			body: metrics.Metrics{
				ID:    "Alloc",
				MType: "gauge",
			},
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:    "Counter not found",
			handler: New(),
			body: metrics.Metrics{
				ID:    "PollCount",
				MType: "counter",
			},
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name: "Gauge ok",
			handler: New(
				WithFileStorer(filestore.New(filestore.WithStorage(
					storage.New(storage.WithGauges(map[string]metrics.Gauge{"Alloc": 1221.23})),
				))),
			),
			body: metrics.Metrics{
				ID:    "Alloc",
				MType: "gauge",
			},
			want: want{
				statusCode: http.StatusOK,
				body: myMetrics{
					ID:    "Alloc",
					MType: "gauge",
					Value: 1221.23,
				},
			},
		},
		{
			name: "Hashed gauge ok",
			handler: New(
				WithFileStorer(filestore.New(filestore.WithStorage(
					storage.New(storage.WithGauges(map[string]metrics.Gauge{"Alloc": 1221.23})),
				))),
				WithKey(key),
			),
			body: metrics.Metrics{
				ID:    "Alloc",
				MType: "gauge",
			},
			want: want{
				statusCode: http.StatusOK,
				body: myMetrics{
					ID:    "Alloc",
					MType: "gauge",
					Value: 1221.23,
				},
			},
		},
		{
			name: "Counter ok",
			handler: New(
				WithFileStorer(filestore.New(filestore.WithStorage(
					storage.New(storage.WithCounters(map[string]metrics.Counter{"PollCount": 42})),
				))),
			),
			body: metrics.Metrics{
				ID:    "PollCount",
				MType: "counter",
			},
			want: want{
				statusCode: http.StatusOK,
				body: myMetrics{
					ID:    "PollCount",
					MType: "counter",
					Delta: 42,
				},
			},
		},
		{
			name: "Hash counter ok",
			handler: New(
				WithFileStorer(filestore.New(filestore.WithStorage(
					storage.New(storage.WithCounters(map[string]metrics.Counter{"PollCount": 42})),
				))),
				WithKey(key),
			),
			body: metrics.Metrics{
				ID:    "PollCount",
				MType: "counter",
			},
			want: want{
				statusCode: http.StatusOK,
				body: myMetrics{
					ID:    "PollCount",
					MType: "counter",
					Delta: 42,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler.router)
			defer ts.Close()

			m := myMetrics{}
			client := resty.New()
			resp, err := client.R().
				SetHeader("Accept", applicationJSON).
				SetHeader("Content-Type", applicationJSON).
				SetBody(tt.body).
				SetResult(&m).
				Post(ts.URL + "/value/")

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
			assert.Equal(t, applicationJSON, resp.Header().Get("Content-Type"))
			assert.EqualValues(t, tt.want.body, m)
		})
	}
}
