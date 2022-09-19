package filestore

import (
	"encoding/json"
	"fmt"
	metricserErrors "github.com/sergeysynergy/metricser/internal/errors"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"os"
)

// ReadMetrics Считывает все метрики из файла.
func (fs *FileStore) ReadMetrics() (*metrics.ProxyMetrics, error) {
	if !fs.restore {
		return nil, metricserErrors.ErrEmptyFilename
	}

	data, err := os.ReadFile(fs.storeFile)
	if err != nil {
		return nil, fs.removeBrokenFile(err)
	}

	prm := metrics.NewProxyMetrics()
	err = json.Unmarshal(data, &prm)
	if err != nil {
		return nil, fs.removeBrokenFile(err)
	}

	if len(prm.Gauges) == 0 && len(prm.Counters) == 0 {
		err = fmt.Errorf("metrics not found in file '%s'", fs.storeFile)
		return nil, fs.removeBrokenFile(err)
	}

	return prm, nil
}
