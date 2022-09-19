package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/sergeysynergy/metricser/internal/data/repository/memory"
	"github.com/sergeysynergy/metricser/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sergeysynergy/metricser/pkg/metrics"
)

func TestUpdateHardBody(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name        string
		contentType string
		body        []byte
		want        want
	}{
		{
			name: "Update unsupported media type",
			body: []byte(""),
			want: want{
				statusCode: http.StatusUnsupportedMediaType,
			},
		},
		{
			name:        "Update not acceptable",
			body:        []byte("Not acceptable"),
			contentType: applicationJSON,
			want: want{
				statusCode: http.StatusNotAcceptable,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(storage.New(memory.New(), nil))
			ts := httptest.NewServer(handler.router)
			defer ts.Close()

			client := resty.New()
			resp, err := client.R().
				EnableTrace().
				SetHeader("Content-type", tt.contentType).
				SetBody(tt.body).
				Post(ts.URL + "/update/")

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
			assert.Equal(t, applicationJSON, resp.Header().Get("Content-Type"))
		})
	}
}

func compressMetrics(t *testing.T, m metrics.Metrics) []byte {
	data, _ := json.Marshal(m)

	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, gzip.BestSpeed)
	if err != nil {
		t.Errorf("failed init compress writer: %v", err)
		return nil
	}
	_, err = w.Write(data)
	if err != nil {
		t.Errorf("failed write data to compress temporary buffer: %v", err)
		return nil
	}
	// обязательно нужно вызвать метод Close() — в противном случае часть данных
	// может не записаться в буфер b; если нужно выгрузить все упакованные данные
	// в какой-то момент сжатия, используйте метод Flush()
	err = w.Close()
	if err != nil {
		t.Errorf("failed compress data: %v", err)
		return nil
	}
	// переменная b содержит сжатые данные
	return b.Bytes()
}

func TestUpdate(t *testing.T) {
	key := "Passw0rd"
	type want struct {
		statusCode int
	}
	tests := []struct {
		name string
		body metrics.Metrics
		key  string
		want want
	}{
		{
			name: "Not implemented",
			body: metrics.Metrics{MType: "unknown"},
			want: want{
				statusCode: http.StatusNotImplemented,
			},
		},
		{
			name: "Gauge ok",
			body: metrics.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: func() *float64 { v := 42.24; return &v }(),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Hashed gauge ok",
			body: metrics.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: func() *float64 { v := 42.24; return &v }(),
				Hash:  metrics.GaugeHash(key, "Alloc", 42.24),
			},
			key: key,
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Counter ok",
			body: metrics.Metrics{
				ID:    "PollCount",
				MType: "counter",
				Delta: func() *int64 { d := int64(2); return &d }(),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Hashed counter ok",
			body: metrics.Metrics{
				ID:    "PollCount",
				MType: "counter",
				Delta: func() *int64 { d := int64(2); return &d }(),
				Hash:  metrics.CounterHash(key, "PollCount", 2),
			},
			key: key,
			want: want{
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(storage.New(memory.New(), nil), WithKey(tt.key))
			ts := httptest.NewServer(handler.router)
			defer ts.Close()

			client := resty.New()
			resp, err := client.R().
				EnableTrace().
				SetHeader("Content-Type", applicationJSON).
				SetHeader("Content-Encoding", "gzip").
				SetBody(compressMetrics(t, tt.body)).
				Post(ts.URL + "/update/")

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
		})
	}
}
