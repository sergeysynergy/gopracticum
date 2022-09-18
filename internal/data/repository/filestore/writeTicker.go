package filestore

// WriteTicker Асинхронно записывает метрики в файл с определённым интервалом.
//func (fs *FileStore) WriteTicker() error {
//	// тикер должен работать только когда задано имя файла
//	if fs.storeFile == "" {
//		return fmt.Errorf("empty file name")
//	}
//	// ... и storeInterval больше нуля
//	if fs.storeInterval == 0 {
//		return fmt.Errorf("store interval should be > 0 to start WriteTicker routine")
//	}
//
//	go func() {
//		ticker := time.NewTicker(fs.storeInterval)
//		defer ticker.Stop()
//
//		for {
//			select {
//			case <-ticker.C:
//				_, err := fs.WriteMetrics()
//				if err != nil {
//					log.Println("[ERROR] Failed to write metrics to disk -", err)
//				}
//			case <-fs.ctx.Done():
//				return
//			}
//		}
//	}()
//
//	return nil
//}
