package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/sergeysynergy/gopracticum/internal/storage"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPost(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "Gauge ok",
			request: "/update/gauge/Alloc/65637.019",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:    "Counter ok",
			request: "/update/counter/PollCount/1",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:    "Unknown metric",
			request: "/update/unknown/testCounter/100",
			want: want{
				statusCode: http.StatusNotImplemented,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New()
			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", handler.Post)
			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, _ := testRequest(t, ts, http.MethodPost, tt.request)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			// повтороне закрытие тела ответа нужно чтобы автотест пропустил
			resp.Body.Close()
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestGet(t *testing.T) {
	type want struct {
		statusCode int
		value      string
	}
	tests := []struct {
		name    string
		handler *Handler
		request string
		want
	}{
		{
			name: "Gauge ok",
			handler: New(WithStorage(storage.NewWithGauges(
				map[string]metrics.Gauge{
					"Alloc": 1221.23,
				},
			))),
			request: "/value/gauge/Alloc",
			want: want{
				statusCode: http.StatusOK,
				value:      "1221.23",
			},
		},
		{
			name:    "Gauge not found",
			handler: New(),
			request: "/value/gauge/NotFound",
			want: want{
				statusCode: http.StatusNotFound,
				value:      "gauge metric with key 'NotFound' not found\n",
			},
		},
		{
			name: "Counter ok",
			handler: New(WithStorage(storage.NewWithCounters(
				map[string]metrics.Counter{
					"PollCount": 42,
				},
			))),
			request: "/value/counter/PollCount",
			want: want{
				statusCode: http.StatusOK,
				value:      "42",
			},
		},
		{
			name:    "Counter not found",
			handler: New(),
			request: "/value/counter/NotFound",
			want: want{
				statusCode: http.StatusNotFound,
				value:      "counter metric with key 'NotFound' not found\n",
			},
		},
		{
			name:    "Not implemented",
			handler: New(),
			request: "/value/not/implemented",
			want: want{
				statusCode: http.StatusNotImplemented,
				value:      "not implemented\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get("/value/{type}/{name}", tt.handler.Get)
			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, body := testRequest(t, ts, http.MethodGet, tt.request)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.value, body)

			// повтороне закрытие тела ответа нужно чтобы автотест пропустил
			resp.Body.Close()
		})
	}
}

func TestList(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "Home ok",
			request: "/",
			want: want{
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New()
			r := chi.NewRouter()
			r.Get("/", handler.List)
			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, _ := testRequest(t, ts, http.MethodGet, tt.request)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			// повтороне закрытие тела ответа нужно чтобы автотест пропустил
			resp.Body.Close()
		})
	}
}

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
			r := chi.NewRouter()
			r.Post("/update/", handler.Update)
			ts := httptest.NewServer(r)
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
			r := chi.NewRouter()
			r.Post("/update/", handler.Update)
			ts := httptest.NewServer(r)
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

func TestValueContentType(t *testing.T) {
	h := New()
	ts := httptest.NewServer(h.router)
	defer ts.Close()

	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "bad type").
		Post(ts.URL + "/value/")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode())
	assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
}

func TestValueUnmarshalError(t *testing.T) {
	h := New()
	ts := httptest.NewServer(h.router)
	defer ts.Close()

	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetBody("{bad bad json").
		Post(ts.URL + "/value/")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotAcceptable, resp.StatusCode())
	assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
}

func TestValue(t *testing.T) {
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
			handler: New(WithStorage(storage.NewWithGauges(
				map[string]metrics.Gauge{
					"Alloc": 1221.23,
				},
			))),
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
			handler: New(WithStorage(storage.NewWithCounters(
				map[string]metrics.Counter{
					"PollCount": 42,
				},
			))),
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
				SetHeader("Accept", "application/json").
				SetHeader("Content-Type", "application/json").
				SetBody(tt.body).
				SetResult(&m).
				Post(ts.URL + "/value/")

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
			assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
			assert.EqualValues(t, tt.want.body, m)
		})
	}
}
