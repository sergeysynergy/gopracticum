package storage

// Shutdown Штатно завершает работу сервиса:
// сохраняет перед выходом значения метрик в файл.
func (s *Storage) Shutdown() error {
	defer s.cancel()

	// Запишем значения метрик в файл.
	prm, err := s.repoDB.GetMetrics()
	if err != nil {
		return err
	}
	err = s.repoFile.WriteMetrics(prm)
	if err != nil {
		return err
	}

	// Штатно завершим работу БД.
	err = s.repoDB.Shutdown()
	if err != nil {
		return err
	}

	return nil
}
