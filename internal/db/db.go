package db

import (
	"context"
	"database/sql"
	"github.com/sergeysynergy/gopracticum/pkg/metrics"
	"log"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/sergeysynergy/gopracticum/internal/storage"
)

const (
	initTimeOut = 60 * time.Second
	//queryTimeOut = 60 * time.Second

	queryCreateTable = `
		CREATE TABLE public.metrics (
			id text NOT NULL,
			type text NOT NULL, 
			value double precision,
			delta bigint,
			PRIMARY KEY (id)
		);
	`
	queryInsertGauge = `INSERT INTO metrics (id, type, value) VALUES ($1, 'gauge', $2)`
	//queryInsertCounter = `INSERT INTO metrics (id, type, delta) VALUES ($1, 'counter', $2)`
	queryUpdateGauge = `UPDATE metrics SET value = $2 WHERE id = $1`
	//queryUpdateCounter = `UPDATE metrics SET delta = $2 WHERE id = $1`
	queryGet        = `SELECT id, type, value, delta FROM metrics WHERE id=$1`
	queryGetMetrics = `SELECT id, type, value, delta FROM metrics`
)

type metricsDB struct {
	ID    string
	MType string
	Value sql.NullFloat64
	Delta sql.NullInt64
}

type Storage struct {
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

func New(dsn string) storage.DBStorer {
	// вернём nil в случае пустой строки DSN
	if dsn == "" {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &Storage{
		ctx:    ctx,
		cancel: cancel,
	}

	// проинициализируем подключение к БД
	err := s.init(dsn)
	if err != nil {
		log.Fatal("[FATAL] Database initialization failed - ", err)
	}

	// опишем стейтменты
	err = s.initStatements()
	if err != nil {
		log.Fatal("[FATAL] Database statements initialization failed - ", err)
	}

	return s
}

func (s *Storage) init(dsn string) error {
	db, err := sql.Open("pgx", dsn)
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

	db.SetMaxOpenConns(40)
	db.SetMaxIdleConns(20)
	db.SetConnMaxIdleTime(time.Second * 60)

	return nil
}

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

func (s *Storage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

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

// получим текущее значение счётчика
func (s *Storage) checkCounter(id string) (metrics.Counter, error) {
	mdb := metricsDB{}
	row := s.stmtCounterGet.QueryRowContext(s.ctx, id)

	err := row.Scan(&mdb.ID, &mdb.MType, &mdb.Value, &mdb.Delta)
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

	return metrics.Counter(mdb.Delta.Int64), nil
}
