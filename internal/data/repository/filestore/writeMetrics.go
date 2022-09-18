package filestore

import (
	"encoding/json"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
	"os"
)

// WriteMetrics Записывает показатели всех метрик в файл в JSON-формате.
func (fs *FileStore) WriteMetrics(prm *metrics.ProxyMetrics) error {
	data, err := json.Marshal(&prm)
	if err != nil {
		return err
	}

	err = os.WriteFile(fs.storeFile, data, 0777)
	if err != nil {
		return err
	}

	log.Printf("written metrics to file '%s': gauges %d, counters %d", fs.storeFile, len(prm.Gauges), len(prm.Counters))
	return nil
}
