package storage

// Put Записывает значение метрики в хранилище Storage для заданного ID.
func (s *Storage) Put(id string, metric interface{}) error {
	return s.repo.Put(id, metric)
}
