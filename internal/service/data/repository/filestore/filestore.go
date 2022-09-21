// Package filestore Пакет предназначен для записи/извлечения значений метрик в файл.
package filestore

// FileStore содержит реализацию репозитория работы с БД, контекст выполнения;
// реализует возможность записи и извлечения значений всех метрик из файла.
type FileStore struct {
	storeFile    string // Имя файла, где хранятся значения метрик (пустое значение — отключает функцию записи на диск).
	removeBroken bool
}

type Options func(fs *FileStore)

// New Создаёт новый объект файлового хранилища FileStorer.
func New(opts ...Options) *FileStore {
	const (
		defaultStoreFile    = "/tmp/devops-metrics-pgsql.json"
		defaultRemoveBroken = false
	)

	fs := &FileStore{
		storeFile:    defaultStoreFile,
		removeBroken: defaultRemoveBroken,
	}
	for _, opt := range opts {
		opt(fs)
	}

	// вернём nil в случае пустого имени файла
	if fs.storeFile == "" {
		return nil
	}

	return fs
}

// WithStoreFile Использует переданный путь к файлу.
func WithStoreFile(filename string) Options {
	return func(fs *FileStore) {
		fs.storeFile = filename
	}
}
