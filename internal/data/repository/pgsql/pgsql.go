// Package pgsql Пакет предназначен для записи значений метрик в базу данных Postgres.
package pgsql

import (
	"context"
	"database/sql"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	"log"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/sergeysynergy/metricser/internal/data/model"
	"github.com/sergeysynergy/metricser/internal/storage"
)

// Storage хранит подключение к БД, контекст выполнения и список SQL-утверждений.
type Storage struct {
	onceDB sync.Once
	db     *sql.DB
	ctx    context.Context
	cancel context.CancelFunc

	stmtGaugeInsert   *sql.Stmt
	stmtCounterInsert *sql.Stmt
	stmtGaugeUpdate   *sql.Stmt
	stmtCounterUpdate *sql.Stmt
	stmtGaugeGet      *sql.Stmt
	stmtCounterGet    *sql.Stmt
	stmtAllUpdate     *sql.Stmt
	stmtAllSelect     *sql.Stmt
}

// New создаёт и инициализирует новую структуру типа Storage.
func New(dsn string) storage.DBStorer {
	// Вернём nil в случае пустой строки DSN. Важно: если не возвращать `nil`, не пройдут автотесты.
	if dsn == "" {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &Storage{
		ctx:    ctx,
		cancel: cancel,
	}

	s.initDB(dsn)

	err := s.initTable()
	if err != nil {
		log.Fatal("[FATAL] Database initialization failed - ", err)
	}

	err = s.initStatements()
	if err != nil {
		log.Fatal("[FATAL] Database statements initialization failed - ", err)
	}

	return s
}

// initDB Инициализирует и настраивает подключение к БД, создаёт таблицу с метриками при необходимости.
func (s *Storage) initDB(dsn string) {
	s.onceDB.Do(func() {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			log.Fatal("[FATAL] Database initialization failed - ", err)
		}
		s.db = db

		db.SetMaxOpenConns(40)
		db.SetMaxIdleConns(20)
		db.SetConnMaxIdleTime(time.Second * 60)
	})
}

// initTable Создаёт в БД таблицу с метриками, если она отсутствует.
func (s *Storage) initTable() error {
	_, err := s.db.ExecContext(s.ctx, "select * from metrics;")
	if err != nil {
		_, err = s.db.ExecContext(s.ctx, `
			CREATE TABLE public.metrics (
				id text NOT NULL,
				type text NOT NULL, 
				value double precision,
				delta bigint,
				PRIMARY KEY (id)
			);
		`)
		if err != nil {
			return err
		}

		log.Println("table `metrics` created")
	}

	return nil
}

// closeStatements Инициализирует SLQ-утверждения при запуске.
func (s *Storage) initStatements() error {
	var err error

	s.stmtGaugeInsert, err = s.db.PrepareContext(s.ctx, "INSERT INTO metrics (id, type, value) VALUES ($1, 'gauge', $2)")
	if err != nil {
		return err
	}

	s.stmtCounterInsert, err = s.db.PrepareContext(s.ctx, "INSERT INTO metrics (id, type, delta) VALUES ($1, 'counter', $2)")
	if err != nil {
		return err
	}

	s.stmtGaugeUpdate, err = s.db.PrepareContext(s.ctx, "UPDATE metrics SET value = $2 WHERE id = $1")
	if err != nil {
		return err
	}

	s.stmtCounterUpdate, err = s.db.PrepareContext(s.ctx, "UPDATE metrics SET delta = $2 WHERE id = $1")
	if err != nil {
		return err
	}

	s.stmtGaugeGet, err = s.db.PrepareContext(s.ctx, "INSERT INTO metrics (id, type, value) VALUES ($1, 'gauge', $2)")
	if err != nil {
		return err
	}

	s.stmtCounterGet, err = s.db.PrepareContext(s.ctx, "SELECT id, type, value, delta FROM metrics WHERE id=$1")
	if err != nil {
		return err
	}

	s.stmtAllUpdate, err = s.db.PrepareContext(s.ctx, "UPDATE metrics SET delta=$2, value=$3 WHERE id = $1")
	if err != nil {
		return err
	}

	s.stmtAllSelect, err = s.db.PrepareContext(s.ctx, "SELECT id, type, value, delta FROM metrics")
	if err != nil {
		return err
	}

	return nil
}

// closeStatements Закрывает SLQ-утверждения при завершении работы.
func (s *Storage) closeStatements() error {
	err := s.stmtGaugeInsert.Close()
	if err != nil {
		return err
	}

	err = s.stmtCounterInsert.Close()
	if err != nil {
		return err
	}

	err = s.stmtGaugeUpdate.Close()
	if err != nil {
		return err
	}

	err = s.stmtCounterUpdate.Close()
	if err != nil {
		return err
	}

	err = s.stmtGaugeGet.Close()
	if err != nil {
		return err
	}

	err = s.stmtCounterGet.Close()
	if err != nil {
		return err
	}

	err = s.stmtAllUpdate.Close()
	if err != nil {
		return err
	}

	err = s.stmtAllSelect.Close()
	if err != nil {
		return err
	}

	return nil
}

// Ping позволяет проверить наличие связи с БД.
func (s *Storage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

// Shutdown штатно завершает работу с БД.
func (s *Storage) Shutdown() error {
	s.cancel()

	err := s.closeStatements()
	if err != nil {
		return err
	}

	err = s.db.Close()
	if err != nil {
		return err
	}

	log.Println("connection to database closed")
	return nil
}

// checkCounter Получает текущее значение счётчика.
func (s *Storage) checkCounter(id string) (metrics.Counter, error) {
	m := model.Metrics{}
	row := s.stmtCounterGet.QueryRowContext(s.ctx, id)

	err := row.Scan(&m.ID, &m.MType, &m.Value, &m.Delta)
	if err == sql.ErrNoRows {
		_, err = s.stmtCounterInsert.ExecContext(s.ctx, id, 0)
		if err != nil {
			return 0, err
		}
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return metrics.Counter(m.Delta.Int64), nil
}
