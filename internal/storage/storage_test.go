package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/sergeysynergy/gopracticum/pkg/metrics"
)

func TestStoragePut(t *testing.T) {
	type want struct {
		wantErr bool
	}
	tests := []struct {
		name  string
		mType string
		ID    string
		value metrics.Gauge
		delta metrics.Counter
		want  want
	}{
		{
			name:  "Gauge ok",
			mType: "gauge",
			ID:    metrics.Alloc,
			value: 1234.42,
		},
		{
			name:  "Counter ok",
			mType: "counter",
			ID:    metrics.PollCount,
			delta: 42,
		},
		{
			name:  "Not implemented",
			mType: "not implemented",
			ID:    "NotImplemented",
			want: want{
				wantErr: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			s := New()

			switch tt.mType {
			case "gauge":
				err = s.Put(tt.ID, tt.value)
			case "counter":
				err = s.Put(tt.ID, tt.delta)
			default:
				err = s.Put(tt.ID, "unknown metric")
			}

			if tt.want.wantErr {
				assert.EqualError(t, err, ErrNotImplemented.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestStorageGet(t *testing.T) {
	type want struct {
		wantErr bool
		value   metrics.Gauge
		delta   metrics.Counter
	}
	tests := []struct {
		name  string
		mType string
		ID    string
		want  want
	}{
		{
			name:  "Gauge ok",
			mType: "gauge",
			ID:    metrics.Alloc,
			want: want{
				value: 1234.42,
			},
		},
		{
			name:  "Counter ok",
			mType: "counter",
			ID:    metrics.PollCount,
			want: want{
				delta: 42,
			},
		},
		{
			name:  "Not found",
			mType: "not found",
			ID:    "NotFound",
			want: want{
				wantErr: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			s := New(
				WithGauges(map[string]metrics.Gauge{metrics.Alloc: 1234.42}),
				WithCounters(map[string]metrics.Counter{metrics.PollCount: 42}),
			)

			val, err := s.Get(tt.ID)
			switch m := val.(type) {
			case metrics.Gauge:
				assert.NoError(t, err)
				assert.Equal(t, m, tt.want.value)
			case metrics.Counter:
				assert.NoError(t, err)
				assert.Equal(t, m, tt.want.delta)
			default:
				assert.EqualError(t, err, ErrNotFound.Error())
			}
		})
	}
}

func TestStoragePutGetMetrics(t *testing.T) {
	type want struct {
		get metrics.ProxyMetric
	}
	tests := []struct {
		name string
		put  metrics.ProxyMetric
		want want
	}{
		{
			name: "Basic put/get",
			put: metrics.ProxyMetric{
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
				Counters: map[string]metrics.Counter{
					"PollCount": 42,
				},
			},
			want: want{
				get: metrics.ProxyMetric{
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
					Counters: map[string]metrics.Counter{
						"PollCount": 42,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			s.PutMetrics(tt.put)
			result := s.GetMetrics()

			assert.EqualValues(t, tt.put, result)
		})
	}
}
