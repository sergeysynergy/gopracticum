package filestore

import (
	"encoding/json"
	"fmt"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
	"os"
)

// ReadMetrics Считывает все метрики из файла.
func (fs *FileStore) ReadMetrics() (prm *metrics.ProxyMetrics, err error) {
	data, err := os.ReadFile(fs.storeFile)
	if err != nil {
		return nil, fs.removeBrokenFile(err)
	}

	err = json.Unmarshal(data, &prm)
	if err != nil {
		return nil, fs.removeBrokenFile(err)
	}

	if len(prm.Gauges) == 0 && len(prm.Counters) == 0 {
		err = fmt.Errorf("metrics not found in file '%s'", fs.storeFile)
		return nil, fs.removeBrokenFile(err)
	}
	log.Println("[DEBUG] Metrics has been read from file:", string(data))

	return
}
