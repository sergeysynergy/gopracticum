// Package pgsql Пакет предназначен для записи/извлечения значений метрик в базу данных Postgres.
package pgsql

import (
	"context"
	"database/sql"
	storage2 "github.com/sergeysynergy/metricser/internal/domain/storage"
	"log"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

// Repo Реализует репозиторий для работы с СУБД Postgres.
type Repo struct {
	onceDB sync.Once
	db     *sql.DB
	ctx    context.Context
	cancel context.CancelFunc
	dsn    string

	stmtGaugeInsert   *sql.Stmt
	stmtCounterInsert *sql.Stmt
	stmtGaugeUpdate   *sql.Stmt
	stmtCounterUpdate *sql.Stmt
	stmtGaugeGet      *sql.Stmt
	stmtCounterGet    *sql.Stmt
	stmtAllUpdate     *sql.Stmt
	stmtAllSelect     *sql.Stmt
}

var _ storage2.RepoDB = new(Repo)

// New создаёт и инициализирует новую структуру типа Storage.
func New(dsn string) *Repo {
	// Вернём nil в случае пустой строки DSN. Важно: если не возвращать `nil`, не пройдут автотесты.
	if dsn == "" {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	r := &Repo{
		ctx:    ctx,
		cancel: cancel,
		dsn:    dsn,
	}

	r.init()

	return r
}

func (r *Repo) init() {
	var err error

	err = r.initDB(r.dsn)
	if err != nil {
		log.Fatal("[FATAL] Database initialization failed - ", err)
	}

	err = r.initTable()
	if err != nil {
		log.Fatal("[FATAL] Database initialization failed - ", err)
	}

	err = r.initStatements()
	if err != nil {
		log.Fatal("[FATAL] Database statements initialization failed - ", err)
	}
}

// initDB Инициализирует и настраивает подключение к БД, создаёт таблицу с метриками при необходимости.
func (r *Repo) initDB(dsn string) (err error) {
	r.onceDB.Do(func() {
		var db *sql.DB
		db, err = sql.Open("pgx", dsn)
		if err != nil {
			return
		}
		r.db = db

		db.SetMaxOpenConns(40)
		db.SetMaxIdleConns(20)
		db.SetConnMaxIdleTime(time.Second * 60)
	})

	return err
}

// initTable Создаёт в БД таблицу с метриками, если она отсутствует.
func (r *Repo) initTable() error {
	_, err := r.db.ExecContext(r.ctx, "select * from metrics;")
	if err != nil {
		_, err = r.db.ExecContext(r.ctx, `
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
func (r *Repo) initStatements() error {
	var err error

	r.stmtGaugeInsert, err = r.db.PrepareContext(r.ctx, "INSERT INTO metrics (id, type, value) VALUES ($1, 'gauge', $2)")
	if err != nil {
		return err
	}

	r.stmtCounterInsert, err = r.db.PrepareContext(r.ctx, "INSERT INTO metrics (id, type, delta) VALUES ($1, 'counter', $2)")
	if err != nil {
		return err
	}

	r.stmtGaugeUpdate, err = r.db.PrepareContext(r.ctx, "UPDATE metrics SET value = $2 WHERE id = $1")
	if err != nil {
		return err
	}

	r.stmtCounterUpdate, err = r.db.PrepareContext(r.ctx, "UPDATE metrics SET delta = $2 WHERE id = $1")
	if err != nil {
		return err
	}

	r.stmtGaugeGet, err = r.db.PrepareContext(r.ctx, "INSERT INTO metrics (id, type, value) VALUES ($1, 'gauge', $2)")
	if err != nil {
		return err
	}

	r.stmtCounterGet, err = r.db.PrepareContext(r.ctx, "SELECT id, type, value, delta FROM metrics WHERE id=$1")
	if err != nil {
		return err
	}

	r.stmtAllUpdate, err = r.db.PrepareContext(r.ctx, "UPDATE metrics SET delta=$2, value=$3 WHERE id = $1")
	if err != nil {
		return err
	}

	r.stmtAllSelect, err = r.db.PrepareContext(r.ctx, "SELECT id, type, value, delta FROM metrics")
	if err != nil {
		return err
	}

	return nil
}
