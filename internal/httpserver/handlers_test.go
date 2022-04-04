package httpserver

import (
	"github.com/go-chi/chi/v5"
	"github.com/sergeysynergy/gopracticum/internal/storage"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerPost(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "Post gauge ok",
			request: "/update/gauge/Alloc/65637.019",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:    "Post counter ok",
			request: "/update/counter/PollCount/1",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:    "Post unknown metric",
			request: "/update/unknown/testCounter/100",
			want: want{
				statusCode: http.StatusNotImplemented,
			},
		},
		{
			name:    "Post not found",
			request: "/update/",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &Handler{
				Storage: storage.New(),
			}
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

func TestHandlerGet(t *testing.T) {
	type want struct {
		statusCode int
		value      string
	}
	tests := []struct {
		name string
		*storage.Storage
		request string
		want
	}{
		{
			name: "Get gauge ok",
			Storage: &storage.Storage{
				Metrics: &metrics.Metrics{
					Gauges: map[metrics.Name]metrics.Gauge{
						"Alloc": 1221.23,
					},
				},
			},
			request: "/value/gauge/Alloc",
			want: want{
				statusCode: http.StatusOK,
				value:      "1221.23",
			},
		},
		{
			name:    "Get gauge not found",
			Storage: &storage.Storage{Metrics: &metrics.Metrics{}},
			request: "/value/gauge/NotFound",
			want: want{
				statusCode: http.StatusNotFound,
				value:      "gauge metric with key 'NotFound' not found\n",
			},
		},
		{
			name: "Get counter ok",
			Storage: &storage.Storage{
				Metrics: &metrics.Metrics{
					Counters: map[metrics.Name]metrics.Counter{
						"PollCount": 42,
					},
				},
			},
			request: "/value/counter/PollCount",
			want: want{
				statusCode: http.StatusOK,
				value:      "42",
			},
		},
		{
			name:    "Get counter not found",
			Storage: &storage.Storage{Metrics: &metrics.Metrics{}},
			request: "/value/counter/NotFound",
			want: want{
				statusCode: http.StatusNotFound,
				value:      "counter metric with key 'NotFound' not found\n",
			},
		},
		{
			name:    "Not implemented",
			Storage: &storage.Storage{},
			request: "/value/not/implemented",
			want: want{
				statusCode: http.StatusNotImplemented,
				value:      "not implemented\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &Handler{
				Storage: tt.Storage,
			}
			r := chi.NewRouter()
			r.Get("/value/{type}/{name}", handler.Get)

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

func TestHandlerList(t *testing.T) {
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
			handler := &Handler{
				Storage: storage.New(),
			}
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
