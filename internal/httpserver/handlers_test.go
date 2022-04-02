package httpserver

import (
	"github.com/go-chi/chi/v5"
	"github.com/sergeysynergy/gopracticum/internal/storage"
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
		value      string
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
