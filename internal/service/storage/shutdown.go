package storage

import "log"

func (s *Storage) Shutdown() error {
	defer s.cancel()

	if s.fileRepo != nil {
		prm, err := s.repo.GetMetrics()
		if err != nil {
			return err
		}

		err = s.fileRepo.JustWriteMetrics(prm)
		if err != nil {
			return err
		}
		log.Println("[DEBUG] Gracefully shutdown filestore")
	}

	if s.repo != nil {
		err := s.repo.Shutdown()
		if err != nil {
			return err
		}
	}

	return nil
}
