package httpserver

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreChecksMiddleware(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name        string
		method      string
		target      string
		contentType string
		want        want
	}{
		{
			name:        "Method not allowed",
			method:      http.MethodGet,
			target:      "/update/gauge/Alloc/1234.5667",
			contentType: "text/plain",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:        "Unprocessable content",
			method:      http.MethodPost,
			target:      "/update/counter/testCounter/42",
			contentType: "application/json",
			want: want{
				statusCode: http.StatusUnprocessableEntity,
			},
		},
		{
			name:        "Status ok",
			method:      http.MethodPost,
			target:      "/update/counter/testCounter/42",
			contentType: "text/plain",
			want: want{
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.target, nil)
			request.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
			})
			preChecksMiddleware(mux).ServeHTTP(w, request)

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
