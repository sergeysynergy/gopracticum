package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/sergeysynergy/gopracticum/internal/filestore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sergeysynergy/gopracticum/internal/storage"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
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
			handler: New(WithRepoStorer(filestore.New(filestore.WithStorage(
				storage.New(storage.WithGauges(map[string]metrics.Gauge{"Alloc": 1221.23})),
			)))),
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
				value:      "storage: metric not found\n",
			},
		},
		{
			name: "Counter ok",
			handler: New(WithRepoStorer(filestore.New(filestore.WithStorage(
				storage.New(storage.WithCounters(map[string]metrics.Counter{"PollCount": 42})),
			)))),
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
				value:      "storage: metric not found\n",
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
