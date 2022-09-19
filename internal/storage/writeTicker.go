package storage

import "fmt"

func (s *Storage) WriteTicker() error {
	if s.fileRepo == nil {
		return fmt.Errorf("empty filestore repository")
	}
	return s.fileRepo.WriteTicker()
}
