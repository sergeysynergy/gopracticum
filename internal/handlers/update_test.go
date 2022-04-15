package handlers

import (
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

func TestUpdate(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name string
		body metrics.Metrics
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New()
			ts := httptest.NewServer(handler.router)
			defer ts.Close()

			client := resty.New()
			resp, err := client.R().
				EnableTrace().
				SetHeader("Content-Type", "application/json").
				SetBody(tt.body).
				Post(ts.URL + "/update/")

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
		})
	}
}

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
			contentType: "application/json",
			want: want{
				statusCode: http.StatusNotAcceptable,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New()
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
			assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
		})
	}
}
