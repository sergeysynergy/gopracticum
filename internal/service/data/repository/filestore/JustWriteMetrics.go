package filestore

import (
	"encoding/json"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"os"
)

// JustWriteMetrics Записывает показатели всех метрик в файл в JSON-формате.
func (fs *FileStore) JustWriteMetrics(prm *metrics.ProxyMetrics) error {
	data, err := json.Marshal(&prm)
	if err != nil {
		return err
	}

	err = os.WriteFile(fs.storeFile, data, 0777)
	if err != nil {
		return err
	}

	return nil
}
