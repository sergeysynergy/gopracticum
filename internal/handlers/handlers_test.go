package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name    string
		request string
		handler interface{}
		want    want
	}{
		{
			name:    "ok list",
			request: "/",
			handler: Handler{},
			want: want{
				statusCode: 200,
			},
		},
		{
			name:    "ok gauge",
			request: "/update/gauge/Alloc/1234.5667",
			handler: Handler{},
			want: want{
				statusCode: 200,
			},
		},
		{
			name:    "ok counter",
			request: "/update/counter/PollCount/1",
			handler: Handler{},
			want: want{
				statusCode: 200,
			},
		},
		{
			name:    "gauge 404",
			handler: Handler{},
			request: "/update/gauge/",
			want: want{
				statusCode: 404,
			},
		},
		{
			name:    "counter 404",
			handler: Handler{},
			request: "/update/counter/",
			want: want{
				statusCode: 404,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()

			//resp, _ := testRequest(t, ts, "GET", "/")
			//assert.Equal(t, http.StatusOK, resp.StatusCode)

			//resp, body = testRequest(t, ts, "GET", "/cars/bmw")
			//assert.Equal(t, http.StatusOK, resp.StatusCode)
			//assert.Equal(t, "brand:bmw", body)
			//
			//resp, body = testRequest(t, ts, "GET", "/cars/bmw")
			//assert.Equal(t, http.StatusOK, resp.StatusCode)
			//assert.Equal(t, "brand:bmw", body)
			//
			//resp, body = testRequest(t, ts, "GET", "/cars/bmw/x5")
			//assert.Equal(t, http.StatusOK, resp.StatusCode)
			//assert.Equal(t, "brand and model:bmw-x5", body)

			/*
				request := httptest.NewRequest(http.MethodPost, tt.request, nil)
				w := httptest.NewRecorder()

				switch tt.handler.(type) {
				case Gauge:
					h := tt.handler.(Gauge)
					h.Storage = storage.New()
					h.PostGauge(w, request)
				case Counter:
					h := tt.handler.(Counter)
					h.Storage = storage.New()
					h.ServeHTTP(w, request)
				}

				result := w.Result()

				assert.Equal(t, tt.want.statusCode, result.StatusCode)

				// Чтобы повторно использовать кешированное TCP-соединение, клиент должен обязательно прочитать
				// тело ответа до конца и закрыть, даже если оно не нужно.
				_, err := ioutil.ReadAll(result.Body)
				require.NoError(t, err)
				err = result.Body.Close()
				require.NoError(t, err)

			*/
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
