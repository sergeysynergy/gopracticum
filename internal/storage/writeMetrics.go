package storage

import "fmt"

func (s *Storage) WriteMetrics() error {
	if s.fileRepo == nil {
		return fmt.Errorf("empty filestore repository")
	}
	return s.fileRepo.WriteMetrics()
}
