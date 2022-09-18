package storage

func (s *Storage) restoreMetrics() (err error) {
	prm, err := s.repoFile.ReadMetrics()
	if err != nil {
		return err
	}

	err = s.PutMetrics(prm)
	if err != nil {
		return err
	}

	return nil
}
