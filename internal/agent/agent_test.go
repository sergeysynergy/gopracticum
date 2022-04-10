package agent

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sergeysynergy/gopracticum/internal/handlers"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAgentSendJsonRequest(t *testing.T) {
	type myMetrics struct {
		ID    string  `json:"id"`
		MType string  `json:"type"`
		Delta int64   `json:"delta,omitempty"`
		Value float64 `json:"value,omitempty"`
	}
	type want struct {
		statusCode int
		metrics    myMetrics
	}
	tests := []struct {
		name    string
		metrics *metrics.Metrics
		want    want
	}{
		{
			name: "gauge ok",
			metrics: &metrics.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: func() *float64 { v := 4242.23; return &v }(),
			},
			want: want{
				statusCode: http.StatusOK,
				metrics: myMetrics{
					ID:    "Alloc",
					MType: "gauge",
					Value: 4242.23,
				},
			},
		},
		{
			name: "counter ok",
			metrics: &metrics.Metrics{
				ID:    "PollCount",
				MType: "counter",
				Delta: func() *int64 { v := int64(2); return &v }(),
			},
			want: want{
				statusCode: http.StatusOK,
				metrics: myMetrics{
					ID:    "PollCount",
					MType: "counter",
					Delta: 2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := handlers.New()
			ts := httptest.NewServer(handler.GetRouter())
			defer ts.Close()

			agent := New(WithAddress(ts.URL[7:]))

			err := agent.sendJSONRequest(context.Background(), tt.metrics)
			assert.NoError(t, err)

			fmt.Println("::", ts.URL+"/value/")

			m := myMetrics{}
			client := resty.New()
			fmt.Println(ts.URL)
			resp, err := client.R().
				SetHeader("Accept", "application/json").
				SetHeader("Content-Type", "application/json").
				SetBody(tt.metrics).
				SetResult(&m).
				Post(ts.URL + "/value/")

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
			assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
			assert.EqualValues(t, tt.want.metrics, m)
		})
	}
}
