package storage

import (
	"fmt"
	serviceErrors "github.com/sergeysynergy/metricser/internal/service/errors"
	"log"
	"time"
)

func (s *Storage) WriteTicker() error {
	if s.fileRepo == nil {
		return serviceErrors.ErrFileStoreNotDefined
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
				err := s.SnapShotCreate()
				if err != nil {
					log.Println("[ERROR] Failed write metrics to disk -", err)
				}
			case <-s.ctx.Done():
				return
			}
		}
	}()

	return nil
}
