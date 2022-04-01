package handlers

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sergeysynergy/gopracticum/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestServeHTTP(t *testing.T) {
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
			name:    "ok gauge",
			request: "/update/gauge/Alloc/1234.5667",
			handler: Gauge{},
			want: want{
				statusCode: 200,
			},
		},
		{
			name:    "ok counter",
			request: "/update/counter/PollCount/1",
			handler: Counter{},
			want: want{
				statusCode: 200,
			},
		},
		{
			name:    "gauge 404",
			handler: Gauge{},
			request: "/update/gauge/",
			want: want{
				statusCode: 404,
			},
		},
		{
			name:    "counter 404",
			handler: Counter{},
			request: "/update/counter/",
			want: want{
				statusCode: 404,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()

			switch tt.handler.(type) {
			case Gauge:
				h := tt.handler.(Gauge)
				h.Storage = storage.New()
				h.ServeHTTP(w, request)
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
		})
	}
}
