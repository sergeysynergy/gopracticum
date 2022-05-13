package filestore

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

func TestFileStoreRestoreMetrics(t *testing.T) {
	type want struct {
		wantErr bool
		body    metrics.ProxyMetrics
	}
	tests := []struct {
		name     string
		fileName string
		body     []byte
		want     want
	}{
		{
			name: "Bad JSON file",
			want: want{
				wantErr: true,
			},
		},
		{
			name: "Metric not found in JSON file",
			body: []byte(`{"just": "json"}`),
			want: want{
				wantErr: true,
			},
		},
		{
			name: "Good JSON file",
			body: []byte(`{"Gauges":{"Alloc":3407240,"BuckHashSys":3972,"Frees":6610,"GCCPUFraction":0.000002760847079840539,"GCSys":4465608,"HeapAlloc":3407240,"HeapIdle":3563520,"HeapInuse":4300800,"HeapObjects":5740,"HeapReleased":3203072,"HeapSys":7864320,"LastGC":1650034139879352300,"Lookups":0,"MCacheInuse":14400,"MCacheSys":16384,"MSpanInuse":68816,"MSpanSys":81920}}`),
			want: want{
				wantErr: false,
				body: metrics.ProxyMetrics{
					Gauges: map[string]metrics.Gauge{
						"Alloc":         3407240,
						"BuckHashSys":   3972,
						"Frees":         6610,
						"GCCPUFraction": 0.000002760847079840539,
						"GCSys":         4465608,
						"HeapAlloc":     3407240,
						"HeapIdle":      3563520,
						"HeapInuse":     4300800,
						"HeapObjects":   5740,
						"HeapReleased":  3203072,
						"HeapSys":       7864320,
						"LastGC":        1650034139879352300,
						"Lookups":       0,
						"MCacheInuse":   14400,
						"MCacheSys":     16384,
						"MSpanInuse":    68816,
						"MSpanSys":      81920,
					},
					Counters: map[string]metrics.Counter{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp("/tmp", "restore-metrics-test*.json")
			assert.NoError(t, err)
			defer os.Remove(f.Name())

			if len(tt.body) > 0 {
				err = os.WriteFile(f.Name(), tt.body, 0777)
				assert.NoError(t, err)
			}

			fs := New(
				WithStoreFile(f.Name()),
				WithRestore(true),
			)
			if !tt.want.wantErr {
				assert.NoError(t, err)
				mcs, _ := fs.GetMetrics()
				assert.Equal(t, tt.want.body, mcs)
			}
		})
	}
}
