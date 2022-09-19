package storage

// Get Извлекает значение метрики из хранилища Storage для заданного ID.
func (s *Storage) Get(id string) (interface{}, error) {
	return s.repo.Get(id)
}
