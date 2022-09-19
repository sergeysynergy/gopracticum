package storage

func (s *Storage) Shutdown() error {
	if s.repo != nil {
		err := s.repo.Shutdown()
		if err != nil {
			return err
		}
	}

	// Shutdown Штатно завершает работу файлового хранилища, сохраняя перед выходом значения метрик в файл.
	if s.fileRepo != nil {
		//defer fs.cancel()
		//err := fs.writeMetrics()
		//if err != nil {
		//	return err
		//}
		prm, err := s.repo.GetMetrics()
		if err != nil {
			return err
		}

		err = s.fileRepo.WriteMetrics(prm)
		if err != nil {
			return err
		}
	}

	return nil
}
