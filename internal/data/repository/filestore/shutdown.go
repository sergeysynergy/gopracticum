package filestore

// Shutdown Штатно завершает работу файлового хранилища, сохраняя перед выходом значения метрик в файл.
//func (fs *FileStore) Shutdown() error {
//	defer fs.cancel()
//
//	_, err := fs.writeMetrics()
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
