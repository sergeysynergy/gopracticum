package storage

import (
	"fmt"
	"log"
	"time"
)

// WriteTicker Асинхронно записывает метрики в файл с определённым интервалом.
func (s *Storage) WriteTicker() error {
	// тикер должен работать только когда задано имя файла для файлового репозитория
	if s.repoFile == nil {
		return fmt.Errorf("empty file repository")
	}
	// ... и storeInterval больше нуля
	if s.storeInterval == 0 {
		return fmt.Errorf("store interval should be > 0 to start WriteTicker routine")
	}

	go func() {
		ticker := time.NewTicker(s.storeInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				prm, err := s.repoDB.GetMetrics()
				if err != nil {
					log.Println("[ERROR] Failed to write metrics to disk -", err)
				}
				err = s.repoFile.WriteMetrics(prm)
				if err != nil {
					log.Println("[ERROR] Failed to write metrics to disk -", err)
				}
			case <-s.ctx.Done():
				return
			}
		}
	}()

	return nil
}
