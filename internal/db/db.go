package db

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/sergeysynergy/gopracticum/internal/storage"
)

const (
	initTimeOut  = 2 * time.Second
	queryTimeOut = 1 * time.Second

	queryCreateTable = `
		CREATE TABLE public.metrics (
			id text NOT NULL,
			type text NOT NULL, 
			value double precision,
			delta bigint,
			PRIMARY KEY (id)
		);
	`
	queryInsertGauge   = `INSERT INTO metrics (id, type, value) VALUES ($1, 'gauge', $2)`
	queryInsertCounter = `INSERT INTO metrics (id, type, delta) VALUES ($1, 'counter', $2)`
	queryUpdateGauge   = `UPDATE metrics SET value = $2 WHERE id = $1`
	queryUpdateCounter = `UPDATE metrics SET delta = $2 WHERE id = $1`
	queryGet           = `SELECT id, type, value, delta FROM metrics WHERE id=$1`
	queryGetMetrics    = `SELECT id, type, value, delta FROM metrics`
)

type metricsDB struct {
	ID    string
	MType string
	Value sql.NullFloat64
	Delta sql.NullInt64
}

type Storage struct {
	db  *sql.DB
	dsn string
}

type Options func(s *Storage)

func New(dsn string, opts ...Options) storage.DBStorer {
	// вернём nil в случае пустой строки DSN
	if dsn == "" {
		return nil
	}

	s := &Storage{
		dsn: dsn,
	}
	for _, opt := range opts {
		opt(s)
	}

	// проинициализируем подключение к БД
	err := s.init()
	if err != nil {
		log.Fatal("[FATAL] Database initialization failed - ", err)
	}

	return s
}

func (s *Storage) init() error {
	db, err := sql.Open("pgx", s.dsn)
	if err != nil {
		return err
	}
	s.db = db

	ctx, cancel := context.WithTimeout(context.Background(), initTimeOut)
	defer cancel()

	_, err = db.ExecContext(ctx, "select * from metrics;")
	if err != nil {
		_, err = db.ExecContext(ctx, queryCreateTable)
		if err != nil {
			return err
		}

		log.Println("table `metrics` created")
	}

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
