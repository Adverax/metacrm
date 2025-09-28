package sql

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IsolationLevel = sql.IsolationLevel

const (
	LevelDefault         = sql.LevelDefault
	LevelReadUncommitted = sql.LevelReadUncommitted
	LevelReadCommitted   = sql.LevelReadCommitted
	LevelWriteCommitted  = sql.LevelWriteCommitted
	LevelRepeatableRead  = sql.LevelRepeatableRead
	LevelSnapshot        = sql.LevelSnapshot
	LevelSerializable    = sql.LevelSerializable
	LevelLinearizable    = sql.LevelLinearizable
)

type ContextWithoutCancel struct {
	context.Context
}

func (that *ContextWithoutCancel) Done() <-chan struct{} {
	return nil
}

func NewContextWithoutCancel(ctx context.Context) context.Context {
	return &ContextWithoutCancel{Context: ctx}
}

type DbId string

type TxOptions = pgx.TxOptions

type RawBytes = sql.RawBytes

type Out = sql.Out

type DBStats = sql.DBStats

type ColumnType struct {
	Name string
	OID  uint32
}

type Scanner interface {
	Scan(src ...interface{}) error
}

type Fetcher func(func(rows Rows) error) error

type Act = func(ctx context.Context) error

type Action func(ctx context.Context, tx Tx) error

var (
	ErrNoRows = sql.ErrNoRows
)

type Manipulator interface {
	Exec(ctx context.Context, args ...interface{}) (Result, error)
	Query(ctx context.Context, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, args ...interface{}) Row
}

type Scope interface {
	DbId(ctx context.Context) DbId
	WithCancel(ctx context.Context) Scope
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	Fetch(ctx context.Context, query string, args ...any) Fetcher
	Begin(ctx context.Context) (Tx, error)
	BeginTx(ctx context.Context, opts *TxOptions) (Tx, error)
}

type DB interface {
	Scope
	Pool() *pgxpool.Pool
	Scope(ctx context.Context) Scope
	Transact(ctx context.Context, action Act) error
	Transaction(ctx context.Context, action Action) error
	TransactionTx(ctx context.Context, action Action, options *TxOptions) error
	InTransaction(ctx context.Context) bool
	WithCancel(ctx context.Context) Scope
	Close()
}

type Tx interface {
	Scope
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Rows interface {
	Scanner
	Err() error
	Next() bool
	Columns() ([]string, error)
	ColumnTypes() ([]*ColumnType, error)
	Close() error
}

type Row interface {
	Scanner
	Err() error
}

type Result interface {
	RowsAffected() (int64, error)
}

type ErrorBuilder interface {
	Build(ctx context.Context, err error, msg string) error
}

type NullString = sql.NullString

type NullInt64 = sql.NullInt64

type NullFloat64 = sql.NullFloat64

type NullTime = sql.NullTime

type NullBool = sql.NullBool
