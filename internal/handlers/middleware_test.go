package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/sergeysynergy/gopracticum/internal/filestore"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sergeysynergy/gopracticum/internal/storage"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

func compress(t *testing.T, data []byte) []byte {
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

func TestGzipDecompressor(t *testing.T) {
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
		want    want
	}{
		{
			name: "Test gzip decompression",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler.router)
			defer ts.Close()

			data, _ := json.Marshal(tt.body)
			m := myMetrics{}
			client := resty.New()

			resp, err := client.R().
				EnableTrace().
				SetHeader("Accept", applicationJSON).
				SetHeader("Accept-Encoding", "gzip").
				SetHeader("Content-Type", applicationJSON).
				SetHeader("Content-Encoding", "gzip").
				SetBody(compress(t, data)).
				SetResult(&m).
				Post(ts.URL + "/value/")

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
			assert.Equal(t, applicationJSON, resp.Header().Get("Content-Type"))
			assert.Equal(t, "gzip", resp.Header().Get("Content-Encoding"))
			assert.EqualValues(t, tt.want.body, m)
		})
	}
}

func TestGzipCompressor(t *testing.T) {
	type want struct {
		statusCode  int
		body        []byte
		contentType string
	}
	tests := []struct {
		name        string
		contentType string
		want        want
	}{
		{
			name:        "Test gzip compression",
			contentType: "html/text",
			want: want{
				statusCode:  http.StatusOK,
				body:        []byte(`<h1>Current metrics data:</h1><div><h2>Gauges</h2><div>Alloc - 1221.23</div></div><div><h2>Counters</h2></div>`),
				contentType: textHTML,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(WithFileStorer(filestore.New(filestore.WithStorage(
				storage.New(storage.WithGauges(map[string]metrics.Gauge{"Alloc": 1221.23})),
			))))
			ts := httptest.NewServer(handler.router)
			defer ts.Close()

			client := resty.New()
			resp, err := client.R().
				EnableTrace().
				SetHeader("Accept-Encoding", "gzip").
				SetHeader("Content-type", tt.contentType).
				Get(ts.URL + "/")

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
			assert.Equal(t, tt.want.contentType, resp.Header().Get("Content-Type"))
			assert.Equal(t, "gzip", resp.Header().Get("Content-Encoding"))
			assert.Equal(t, tt.want.body, resp.Body())
		})
	}
}
