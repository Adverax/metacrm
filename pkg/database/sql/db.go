package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type manipulator interface {
	DoExec(ctx context.Context, args ...interface{}) (Result, error)
	DoQuery(ctx context.Context, args ...interface{}) (Rows, error)
	DoQueryRow(ctx context.Context, args ...interface{}) Row
}

type executor interface {
	WithCancel(ctx context.Context) Scope
	DoExec(ctx context.Context, query string, args ...interface{}) (Result, error)
	DoQuery(ctx context.Context, query string, args ...interface{}) (Rows, error)
	DoQueryRow(ctx context.Context, query string, args ...interface{}) Row
}

type executorEx interface {
	executor
	DbId(ctx context.Context) DbId
	DoBegin(ctx context.Context) (Tx, error)
	DoBeginTx(ctx context.Context, opts *TxOptions) (Tx, error)
}

type ExecAction func(ctx context.Context, query string, args ...interface{}) (Result, error)
type QueryAction func(ctx context.Context, query string, args ...interface{}) (Rows, error)
type QueryRowAction func(ctx context.Context, query string, args ...interface{}) Row
type CommitAction func(ctx context.Context) error
type RollbackAction func(ctx context.Context) error
type BeginAction func(ctx context.Context, opts *TxOptions) (Tx, error)

type Handler interface {
	Query(ctx context.Context, action QueryAction, query string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, action QueryRowAction, query string, args ...interface{}) Row
	Exec(ctx context.Context, action ExecAction, query string, args ...interface{}) (Result, error)
	BeginTx(ctx context.Context, opts *TxOptions, action BeginAction) (Tx, error)
	Commit(ctx context.Context, action CommitAction) error
	Rollback(ctx context.Context, action RollbackAction) error
}

func Register(name string, driver driver.Driver) {
	sql.Register(name, driver)
}

var ErrTxDone = sql.ErrTxDone

type cancelableManipulator struct {
	manipulator
}

func (that *cancelableManipulator) Exec(ctx context.Context, args ...interface{}) (Result, error) {
	return that.manipulator.DoExec(ctx, args...)
}

func (that *cancelableManipulator) Query(ctx context.Context, args ...interface{}) (Rows, error) {
	return that.manipulator.DoQuery(ctx, args...)
}

func (that *cancelableManipulator) QueryRow(ctx context.Context, args ...interface{}) Row {
	return that.manipulator.DoQueryRow(ctx, args...)
}

type cancelableExecutorEx struct {
	executorEx
}

func (that *cancelableExecutorEx) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return that.executorEx.DoExec(ctx, query, args...)
}

func (that *cancelableExecutorEx) Fetch(ctx context.Context, query string, args ...any) Fetcher {
	return newFetcher(ctx, that, query, args...)
}

func (that *cancelableExecutorEx) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	return that.executorEx.DoQuery(ctx, query, args...)
}

func (that *cancelableExecutorEx) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return that.executorEx.DoQueryRow(ctx, query, args...)
}

func (that *cancelableExecutorEx) Begin(ctx context.Context) (Tx, error) {
	return that.executorEx.DoBegin(ctx)
}

func (that *cancelableExecutorEx) BeginTx(ctx context.Context, opts *TxOptions) (Tx, error) {
	return that.executorEx.DoBeginTx(ctx, opts)
}

type db struct {
	pool    *pgxpool.Pool
	dbId    DbId
	errors  ErrorBuilder
	source  string
	handler Handler
}

func (that *db) Pool() *pgxpool.Pool {
	return that.pool
}

func (that *db) Close() {
	that.pool.Close()
}

func (that *db) Source() string {
	return that.source
}

func (that *db) DbId(context.Context) DbId {
	return that.dbId
}

func (that *db) WithCancel(context.Context) Scope {
	return &cancelableExecutorEx{executorEx: that}
}

func (that *db) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return that.DoExec(NewContextWithoutCancel(ctx), query, args...)
}

func (that *db) DoExec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return that.handler.Exec(ctx, func(ctx context.Context, query string, args ...interface{}) (Result, error) {
		res, err := that.pool.Exec(ctx, query, args...)
		if err != nil {
			return nil, that.errors.Build(ctx, err, "DB.ExecContext")
		}

		return &result{res: res, db: that, ctx: ctx}, nil
	}, query, args...)
}

func (that *db) Fetch(ctx context.Context, query string, args ...any) Fetcher {
	return newFetcher(ctx, that, query, args...)
}

func (that *db) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	return that.DoQuery(NewContextWithoutCancel(ctx), query, args...)
}

func (that *db) DoQuery(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	return that.handler.Query(ctx, func(ctx context.Context, query string, args ...interface{}) (Rows, error) {
		rs, err := that.pool.Query(ctx, query, args...)
		if err != nil {
			return nil, that.errors.Build(ctx, err, "DB.QueryContext")
		}

		return &rows{rows: rs, db: that, ctx: ctx}, nil
	}, query, args...)
}

func (that *db) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return that.DoQueryRow(NewContextWithoutCancel(ctx), query, args...)
}

func (that *db) DoQueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return that.handler.QueryRow(ctx, func(ctx context.Context, query string, args ...interface{}) Row {
		r := that.pool.QueryRow(ctx, query, args...)
		return &row{row: r, db: that, ctx: ctx}
	}, query, args...)
}

func (that *db) Begin(ctx context.Context) (Tx, error) {
	return that.DoBegin(NewContextWithoutCancel(ctx))
}

func (that *db) DoBegin(ctx context.Context) (Tx, error) {
	return that.DoBeginTx(ctx, nil)
}

func (that *db) BeginTx(ctx context.Context, _ *TxOptions) (Tx, error) {
	return that.DoBeginTx(NewContextWithoutCancel(ctx), nil)
}

func (that *db) DoBeginTx(ctx context.Context, opts *TxOptions) (Tx, error) {
	return that.handler.BeginTx(ctx, opts, func(ctx context.Context, opts *TxOptions) (Tx, error) {
		if opts == nil {
			opts = &TxOptions{
				AccessMode: pgx.ReadWrite,
			}
		}
		res, err := that.pool.BeginTx(ctx, *opts)
		if err != nil {
			return nil, that.errors.Build(ctx, err, "that.BeginTx")
		}

		return &tx{tx: res, db: that}, nil
	})
}

type tx struct {
	db *db
	tx pgx.Tx
}

func (that *tx) DbId(_ context.Context) DbId {
	return that.db.dbId
}

func (that *tx) WithCancel(_ context.Context) Scope {
	return &cancelableExecutorEx{executorEx: that}
}

func (that *tx) Begin(ctx context.Context) (Tx, error) {
	return that.DoBegin(NewContextWithoutCancel(ctx))
}

func (that *tx) DoBegin(_ context.Context) (Tx, error) {
	return &tx2{tx: that.tx, db: that.db, level: 1}, nil
}

func (that *tx) BeginTx(ctx context.Context, opts *TxOptions) (Tx, error) {
	return that.DoBeginTx(NewContextWithoutCancel(ctx), opts)
}

func (that *tx) DoBeginTx(_ context.Context, _ *TxOptions) (Tx, error) {
	return &tx2{tx: that.tx, db: that.db, level: 1}, nil
}

func (that *tx) Commit(ctx context.Context) error {
	return that.db.handler.Commit(ctx, func(ctx context.Context) error {
		err := that.tx.Commit(ctx)
		if err != nil {
			return that.db.errors.Build(context.Background(), err, "Tx.Commit")
		}
		return nil
	})
}

func (that *tx) Rollback(ctx context.Context) error {
	return that.db.handler.Rollback(ctx, func(ctx context.Context) error {
		err := that.tx.Rollback(ctx)
		if err != nil {
			if errors.Is(err, ErrTxDone) {
				return err
			}
			return that.db.errors.Build(context.Background(), err, "Tx.Rollback")
		}

		return nil
	})
}

func (that *tx) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return that.DoExec(NewContextWithoutCancel(ctx), query, args...)
}

func (that *tx) DoExec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return that.db.handler.Exec(ctx, func(ctx context.Context, query string, args ...interface{}) (Result, error) {
		res, err := that.tx.Exec(ctx, query, args...)
		if err != nil {
			return nil, that.db.errors.Build(ctx, err, "Tx.ExecContext")
		}
		return &result{res: res, db: that.db, ctx: ctx}, nil
	}, query, args...)
}

func (that *tx) Fetch(ctx context.Context, query string, args ...any) Fetcher {
	return newFetcher(ctx, that, query, args...)
}

func (that *tx) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	return that.DoQuery(NewContextWithoutCancel(ctx), query, args...)
}

func (that *tx) DoQuery(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	return that.db.handler.Query(ctx, func(ctx context.Context, query string, args ...interface{}) (Rows, error) {
		rs, err := that.tx.Query(ctx, query, args...)
		if err != nil {
			return nil, that.db.errors.Build(ctx, err, "Tx.QueryContext")
		}

		return &rows{rows: rs, db: that.db, ctx: ctx}, nil
	}, query, args...)
}

func (that *tx) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return that.DoQueryRow(NewContextWithoutCancel(ctx), query, args...)
}

func (that *tx) DoQueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return that.db.handler.QueryRow(ctx, func(ctx context.Context, query string, args ...interface{}) Row {
		r := that.tx.QueryRow(ctx, query, args...)
		return &row{row: r, db: that.db, ctx: ctx}
	}, query, args...)
}

type tx2 struct {
	db    *db
	tx    pgx.Tx
	level int
}

func (that *tx2) DbId(_ context.Context) DbId {
	return that.db.dbId
}

func (that *tx2) WithCancel(_ context.Context) Scope {
	return &cancelableExecutorEx{executorEx: that}
}

func (that *tx2) Begin(ctx context.Context) (Tx, error) {
	return that.DoBegin(ctx)
}

func (that *tx2) DoBegin(_ context.Context) (Tx, error) {
	return &tx2{tx: that.tx, db: that.db, level: that.level + 1}, nil
}

func (that *tx2) BeginTx(ctx context.Context, opts *TxOptions) (Tx, error) {
	return that.DoBeginTx(ctx, opts)
}

func (that *tx2) DoBeginTx(_ context.Context, _ *TxOptions) (Tx, error) {
	return &tx2{tx: that.tx, db: that.db, level: that.level + 1}, nil
}

func (that *tx2) Commit(_ context.Context) error {
	return nil
}

func (that *tx2) Rollback(_ context.Context) error {
	return nil
}

func (that *tx2) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return that.DoExec(ctx, query, args...)
}

func (that *tx2) DoExec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return that.db.handler.Exec(ctx, func(ctx context.Context, query string, args ...interface{}) (Result, error) {
		res, err := that.tx.Exec(ctx, query, args...)
		if err != nil {
			return nil, that.db.errors.Build(ctx, err, "Tx2.ExecContext")
		}
		return &result{res: res, db: that.db, ctx: ctx}, nil
	}, query, args...)
}

func (that *tx2) Fetch(ctx context.Context, query string, args ...any) Fetcher {
	return newFetcher(ctx, that, query, args...)
}

func (that *tx2) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	return that.DoQuery(ctx, query, args...)
}

func (that *tx2) DoQuery(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	return that.db.handler.Query(ctx, func(ctx context.Context, query string, args ...interface{}) (Rows, error) {
		rs, err := that.tx.Query(ctx, query, args...)
		if err != nil {
			return nil, that.db.errors.Build(ctx, err, "Tx2.QueryContext")
		}

		return &rows{rows: rs, db: that.db, ctx: ctx}, nil
	}, query, args...)
}

func (that *tx2) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return that.DoQueryRow(ctx, query, args...)
}

func (that *tx2) DoQueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return that.db.handler.QueryRow(ctx, func(ctx context.Context, query string, args ...interface{}) Row {
		r := that.tx.QueryRow(ctx, query, args...)
		return &row{row: r, db: that.db, ctx: ctx}
	}, query, args...)
}

type rows struct {
	db   *db
	rows pgx.Rows
	ctx  context.Context
}

func (that *rows) Next() bool {
	return that.rows.Next()
}

func (that *rows) Err() error {
	err := that.rows.Err()
	if err != nil {
		return that.db.errors.Build(that.ctx, err, "rows.Err")
	}
	return nil
}

// Columns — возвращает имена колонок
func (that *rows) Columns() ([]string, error) {
	fields := that.rows.FieldDescriptions()
	cols := make([]string, len(fields))
	for i, f := range fields {
		cols[i] = string(f.Name)
	}
	return cols, nil
}

// ColumnTypes — возвращает массив ColumnType
func (that *rows) ColumnTypes() ([]*ColumnType, error) {
	fields := that.rows.FieldDescriptions()
	cols := make([]*ColumnType, len(fields))
	for i, f := range fields {
		cols[i] = &ColumnType{
			Name: string(f.Name),
			OID:  f.DataTypeOID,
		}
	}
	return cols, nil
}

func (that *rows) Scan(dest ...interface{}) error {
	err := that.rows.Scan(dest...)
	if err != nil {
		return that.db.errors.Build(that.ctx, err, "rows.Scan")
	}
	return nil
}

func (that *rows) Close() error {
	that.rows.Close()
	return nil
}

type row struct {
	err error
	row pgx.Row
	db  *db
	ctx context.Context
}

func (that *row) Scan(dest ...interface{}) error {
	if that.err != nil {
		return that.err
	}

	err := that.row.Scan(dest...)
	if err != nil {
		return that.db.errors.Build(that.ctx, err, "rows.Scan")
	}
	return nil
}

func (that *row) Err() error {
	return that.err
}

type result struct {
	db  *db
	res pgconn.CommandTag
	ctx context.Context
}

func (res *result) RowsAffected() (int64, error) {
	return res.res.RowsAffected(), nil
}

type database struct {
	*db
}

func (that *database) DbId(ctx context.Context) DbId {
	return that.db.DbId(ctx)
}

func (that *database) Scope(ctx context.Context) Scope {
	if scope := FromContext(ctx, that.DbId(ctx)); scope != nil {
		return scope
	}

	return that.db
}

func (that *database) Transact(ctx context.Context, action Act) error {
	tx, err := that.Scope(ctx).BeginTx(ctx, &TxOptions{AccessMode: pgx.ReadWrite})
	if err != nil {
		return fmt.Errorf("BeginTx: %w", err)
	}
	ctx2 := ToContext(ctx, tx)
	defer tx.Rollback(ctx2)

	err = action(ctx2)
	if err != nil {
		return fmt.Errorf("action: %w", err)
	}

	return tx.Commit(ctx2)
}

func (that *database) Transaction(
	ctx context.Context,
	action Action,
) (err error) {
	return that.TransactionTx(ctx, action, &TxOptions{AccessMode: pgx.ReadWrite})
}

func (that *database) TransactionTx(
	ctx context.Context,
	action Action,
	options *TxOptions,
) error {
	tx, err := that.Scope(ctx).BeginTx(ctx, options)
	if err != nil {
		return fmt.Errorf("BeginTx: %w", err)
	}
	ctx2 := ToContext(ctx, tx)
	defer tx.Rollback(ctx2)

	err = action(ctx2, tx)
	if err != nil {
		return fmt.Errorf("action: %w", err)
	}

	return tx.Commit(ctx2)
}

func (that *database) InTransaction(
	ctx context.Context,
) bool {
	scope := FromContext(ctx, that.DbId(ctx))
	return scope != nil
}

func (that *database) WithCancel(ctx context.Context) Scope {
	return that.Scope(ctx).WithCancel(ctx)
}

func (that *database) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return that.Scope(ctx).Exec(ctx, query, args...)
}

func (that *database) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	return that.Scope(ctx).Query(ctx, query, args...)
}

func (that *database) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return that.Scope(ctx).QueryRow(ctx, query, args...)
}

func (that *database) Begin(ctx context.Context) (Tx, error) {
	return that.Scope(ctx).Begin(ctx)
}

func (that *database) BeginTx(ctx context.Context, opts *TxOptions) (Tx, error) {
	return that.Scope(ctx).BeginTx(ctx, opts)
}
