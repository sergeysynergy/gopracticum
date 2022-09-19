package storage

func (s *Storage) Shutdown() error {
	if s.repo != nil {
		err := s.repo.Shutdown()
		if err != nil {
			return err
		}
	}

	if s.fileRepo != nil {
		err := s.fileRepo.Shutdown()
		if err != nil {
			return err
		}
	}

	return nil
}
