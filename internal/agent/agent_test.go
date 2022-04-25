package agent

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sergeysynergy/gopracticum/internal/handlers"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

func TestAgentSendJsonRequest(t *testing.T) {
	key := "/tmp/mQUXIrA"
	type myMetrics struct {
		ID    string  `json:"id"`
		MType string  `json:"type"`
		Delta int64   `json:"delta,omitempty"`
		Value float64 `json:"value,omitempty"`
		Hash  string  `json:"hash,omitempty"`
	}
	type want struct {
		wantErr    bool
		statusCode int
		metrics    myMetrics
	}
	tests := []struct {
		name          string
		metricsUpdate *metrics.Metrics
		metricsValue  *metrics.Metrics
		key           string
		want          want
	}{
		{
			name: "gauge ok",
			metricsUpdate: &metrics.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: func() *float64 { v := 4242.23; return &v }(),
			},
			metricsValue: &metrics.Metrics{
				ID:    "Alloc",
				MType: "gauge",
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
			name: "Hash gauge ok",
			metricsUpdate: &metrics.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: func() *float64 { v := 4242.23; return &v }(),
				Hash:  metrics.GaugeHash(key, "Alloc", 4242.23),
			},
			metricsValue: &metrics.Metrics{
				ID:    "Alloc",
				MType: "gauge",
			},
			key: key,
			want: want{
				statusCode: http.StatusOK,
				metrics: myMetrics{
					ID:    "Alloc",
					MType: "gauge",
					Value: 4242.23,
					Hash:  metrics.GaugeHash(key, "Alloc", 4242.23),
				},
			},
		},
		{
			name: "counter ok",
			metricsUpdate: &metrics.Metrics{
				ID:    "PollCount",
				MType: "counter",
				Delta: func() *int64 { v := int64(2); return &v }(),
			},
			metricsValue: &metrics.Metrics{
				ID:    "PollCount",
				MType: "counter",
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
		{
			name: "Hash counter ok",
			metricsUpdate: &metrics.Metrics{
				ID:    "PollCount",
				MType: "counter",
				Delta: func() *int64 { v := int64(2); return &v }(),
				Hash:  metrics.CounterHash(key, "PollCount", 2),
			},
			metricsValue: &metrics.Metrics{
				ID:    "PollCount",
				MType: "counter",
			},
			key: key,
			want: want{
				statusCode: http.StatusOK,
				metrics: myMetrics{
					ID:    "PollCount",
					MType: "counter",
					Delta: 2,
					Hash:  metrics.CounterHash(key, "PollCount", 2),
				},
			},
		},
		{
			name: "Hash counter bad",
			metricsUpdate: &metrics.Metrics{
				ID:    "PollCount",
				MType: "counter",
				Delta: func() *int64 { v := int64(2); return &v }(),
				Hash:  "bad hash",
			},
			metricsValue: &metrics.Metrics{
				ID:    "PollCount",
				MType: "counter",
			},
			key: key,
			want: want{
				wantErr:    true,
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := handlers.New(handlers.WithKey(tt.key))
			ts := httptest.NewServer(handler.GetRouter())
			defer ts.Close()

			agent := New(
				WithAddress(ts.URL[7:]),
				WithKey(tt.key),
			)

			resp, err := agent.sendJSONRequest(context.Background(), tt.metricsUpdate)

			if tt.want.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.want.statusCode, resp.StatusCode())
				return
			}

			assert.NoError(t, err)

			m := myMetrics{}
			client := resty.New()
			fmt.Println(ts.URL)
			resp, err = client.R().
				SetHeader("Accept", "application/json").
				SetHeader("Accept-Encoding", "gzip").
				SetHeader("Content-Type", "application/json").
				SetBody(tt.metricsValue).
				SetResult(&m).
				Post(ts.URL + "/value/")

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
			assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
			assert.EqualValues(t, tt.want.metrics, m)
		})
	}
}
