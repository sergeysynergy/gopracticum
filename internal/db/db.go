package db

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/sergeysynergy/gopracticum/internal/storage"
)

type Storage struct {
	*storage.Storage
	db  *sql.DB
	dsn string
}

type Options func(s *Storage)

func New(opts ...Options) storage.DBStorer {
	defaultDSN := "user=" + os.Getenv("USER") + " password=Passw0rd33 host=localhost port=5432 dbname=metrics"

	s := &Storage{
		Storage: storage.New(),
		dsn:     defaultDSN,
	}
	for _, opt := range opts {
		opt(s)
	}

	// вернём nil в случае пустой строки DSN
	if s.dsn == "" {
		return nil
	}

	// создаём Storage, если он не был проинициализирован через WithStorage
	if s.Storage == nil {
		s.Storage = storage.New()
	}

	// проинициализируем подключение к БД
	err := s.init()
	if err != nil {
		log.Fatal("[FATAL] Database initialization failed - ", err)
	}

	return s
}

func WithStorage(st *storage.Storage) Options {
	return func(s *Storage) {
		s.Storage = st
	}
}

func WithDSN(dsn string) Options {
	return func(s *Storage) {
		if dsn != "" {
			s.dsn = dsn
		}
	}
}

func (s *Storage) init() error {
	db, err := sql.Open("pgx", s.dsn)
	if err != nil {
		return err
	}
	s.db = db

	return nil
}

func (s *Storage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Shutdown() error {
	err := s.db.Close()
	if err != nil {
		return err
	}

	log.Println("connection to database closed")
	return nil
}
