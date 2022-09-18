// Package filestore Пакет предназначен для записи/извлечения значений метрик в файл.
package filestore

import (
	"github.com/sergeysynergy/metricser/internal/domain/storage"
	"log"
)

// FileStore содержит реализацию репозитория работы с БД, контекст выполнения;
// реализует возможность записи и извлечения значений всех метрик из файла.
type FileStore struct {
	//storage2.Repo
	//ctx           context.Context
	//cancel        context.CancelFunc
	storeFile string // Имя файла, где хранятся значения метрик (пустое значение — отключает функцию записи на диск).
	//storeInterval time.Duration // интервал периодического сохранения метрик на диск, 0 — делает запись синхронной
	//restore bool
	removeBroken bool
}

var _ storage.RepoFile = new(FileStore)

type Options func(fs *FileStore)

// New Создаёт новый объект файлового хранилища FileStorer.
func New(opts ...Options) storage.RepoFile {
	const (
		defaultStoreFile = "/tmp/devops-metrics-pgsql.json"
		//defaultRestore       = false
		//defaultStoreInterval = 300 * time.Second
		defaultRemoveBroken = false
	)

	//ctx, cancel := context.WithCancel(context.Background())

	fs := &FileStore{
		//ctx:       ctx,
		//cancel:    cancel,
		storeFile: defaultStoreFile,
		//restore:       defaultRestore,
		//storeInterval: defaultStoreInterval,
		removeBroken: defaultRemoveBroken,
	}
	for _, opt := range opts {
		opt(fs)
	}

	// вернём nil в случае пустого имени файла
	//if fs.storeFile == "" {
	//	return nil
	//}

	// создаём Storer, если он не был проинициализирован через WithStorer
	//if fs.Repo == nil {
	//	fs.Repo = storage2.New()
	//}

	// проинициализируем файловое хранилище
	//err := fs.init()
	//if err != nil {
	//	log.Fatal("[FATAL] File store initialization failed - ", err)
	//}

	if fs.storeFile == "" {
		log.Fatal("[FATAL] File store initialization failed - empty file name")
	}

	return fs
}

// WithStorer Использует переданный репозиторий.
//func WithStorer(st storage2.Repo) Options {
//	return func(fs *FileStore) {
//		if st != nil {
//			fs.Repo = st
//		}
//	}
//}

// WithRestore Определяет флаг, нужно ли восстанавливать при запуске значения метрик из файла.
//func WithRestore(restore bool) Options {
//	return func(fs *FileStore) {
//		fs.restore = restore
//	}
//}

// WithStoreFile Использует переданный путь к файлу.
func WithStoreFile(filename string) Options {
	return func(fs *FileStore) {
		fs.storeFile = filename
	}
}

//// WithStoreInterval Определяет интервал сохранения метрик в файл.
//func WithStoreInterval(interval time.Duration) Options {
//	return func(fs *FileStore) {
//		fs.storeInterval = interval
//	}
//}

// Init Выполним инициализацию файлового хранилища.
//func (fs *FileStore) init() error {
//	if fs.storeFile == "" {
//		return fmt.Errorf("empty file name")
//	}

//err := fs.restoreMetrics()
//if err != nil {
//	log.Printf("[WARNING] Failed to restore metrics from file '%s' - %s\n", fs.storeFile, err)
//}

//	return nil
//}
