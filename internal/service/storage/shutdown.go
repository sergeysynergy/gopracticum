package storage

func (s *Storage) Shutdown() error {
	if s.fileRepo != nil {
		//err := s.fileRepo.Shutdown()

		//defer fs.cancel()
		prm, err := s.repo.GetMetrics()
		if err != nil {
			return err
		}

		err = s.fileRepo.JustWriteMetrics(prm)
		if err != nil {
			return err
		}
	}

	if s.repo != nil {
		err := s.repo.Shutdown()
		if err != nil {
			return err
		}
	}

	return nil
}
