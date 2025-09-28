package sql

import (
	"context"
	"time"
)

type Middleware func(handler Handler) Handler

type DummyHandler struct{}

func (that *DummyHandler) Query(ctx context.Context, action QueryAction, query string, args ...interface{}) (Rows, error) {
	return action(ctx, query, args...)
}

func (that *DummyHandler) QueryRow(ctx context.Context, action QueryRowAction, query string, args ...interface{}) Row {
	return action(ctx, query, args...)
}

func (that *DummyHandler) Exec(ctx context.Context, action ExecAction, query string, args ...interface{}) (Result, error) {
	return action(ctx, query, args...)
}

func (that *DummyHandler) BeginTx(ctx context.Context, opts *TxOptions, action BeginAction) (Tx, error) {
	return action(ctx, opts)
}

func (that *DummyHandler) Rollback(ctx context.Context, action RollbackAction) error {
	return action(ctx)
}

func (that *DummyHandler) Commit(ctx context.Context, action CommitAction) error {
	return action(ctx)
}

func NewDummyHandler() *DummyHandler {
	return &DummyHandler{}
}

type CustomHandler struct {
	behavior Behavior
	next     Handler
}

func (that *CustomHandler) Query(ctx context.Context, action QueryAction, query string, args ...interface{}) (rows Rows, err error) {
	err = that.behavior.Apply(ctx, func(ctx context.Context) error {
		rows, err = that.next.Query(ctx, action, query, args...)
		return err
	}, query, args...)
	return
}

func (that *CustomHandler) QueryRow(ctx context.Context, action QueryRowAction, query string, args ...interface{}) (row Row) {
	_ = that.behavior.Apply(ctx, func(ctx context.Context) error {
		row = that.next.QueryRow(ctx, action, query, args...)
		return nil
	}, query, args...)
	return
}

func (that *CustomHandler) Exec(ctx context.Context, action ExecAction, query string, args ...interface{}) (res Result, err error) {
	err = that.behavior.Apply(ctx, func(ctx context.Context) error {
		res, err = that.next.Exec(ctx, action, query, args...)
		return err
	}, query, args...)
	return
}

func (that *CustomHandler) BeginTx(ctx context.Context, opts *TxOptions, action BeginAction) (tx Tx, err error) {
	err = that.behavior.Apply(
		ctx,
		func(ctx context.Context) error {
			tx, err = that.next.BeginTx(ctx, opts, action)
			return err
		},
		"BEGIN TRANSACTION",
	)
	return
}

func (that *CustomHandler) Rollback(ctx context.Context, action RollbackAction) error {
	return that.behavior.Apply(
		ctx,
		func(ctx context.Context) error {
			return that.next.Rollback(ctx, action)
		},
		"ROLLBACK",
	)
}

func (that *CustomHandler) Commit(ctx context.Context, action CommitAction) error {
	return that.behavior.Apply(
		ctx,
		func(ctx context.Context) error {
			return that.next.Commit(ctx, action)
		},
		"COMMIT",
	)
}

func NewCustomHandler(behavior Behavior, next Handler) *CustomHandler {
	return &CustomHandler{
		behavior: behavior,
		next:     next,
	}
}

type Behavior interface {
	Apply(
		ctx context.Context,
		action func(ctx context.Context) error,
		query string,
		args ...interface{},
	) error
}

type DummyBehavior struct{}

func (that *DummyBehavior) Apply(
	ctx context.Context,
	action func(ctx context.Context) error,
	query string,
	args ...interface{},
) error {
	return action(ctx)
}

type accumulator interface {
	Add(duration time.Duration)
}
