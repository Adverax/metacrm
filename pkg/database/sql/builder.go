package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/adverax/metacrm/pkg/core"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const DatabaseComponentName = "database"

type Builder struct {
	db              *db
	dsn             DSN
	middlewares     []Middleware
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
	tracer          pgx.QueryTracer
	beforeAcquire   func(ctx context.Context, conn *pgx.Conn) bool
	errs            core.Errors
}

func NewBuilder() *Builder {
	return &Builder{
		db: &db{
			dbId: "platform",
		},
		dsn:             DefaultDSN(),
		maxOpenConns:    5,
		maxIdleConns:    5,
		connMaxLifetime: 5 * time.Minute,
		connMaxIdleTime: 5 * time.Minute,
	}
}

func (that *Builder) WithDatabaseID(id DbId) *Builder {
	that.db.dbId = id
	return that
}

func (that *Builder) WithDSN(dsn string) *Builder {
	cfg, err := pgconn.ParseConfig(dsn)
	if err != nil {
		that.errs.AddError(err)
		return that
	}

	that.dsn = DSN{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Database: cfg.Database,
	}
	return that
}

func (that *Builder) WithErrorBuilder(errors ErrorBuilder) *Builder {
	that.db.errors = errors
	return that
}

func (that *Builder) WithHost(host string) *Builder {
	that.dsn.Host = host
	return that
}

func (that *Builder) WithPort(port uint16) *Builder {
	that.dsn.Port = port
	return that
}

func (that *Builder) WithUser(user string) *Builder {
	that.dsn.User = user
	return that
}

func (that *Builder) WithPassword(password string) *Builder {
	that.dsn.Password = password
	return that
}

func (that *Builder) WithDatabase(database string) *Builder {
	that.dsn.Database = database
	return that
}

func (that *Builder) WithMaxOpenConns(max int) *Builder {
	that.maxOpenConns = max
	return that
}

func (that *Builder) WithMaxIdleConns(max int) *Builder {
	that.maxIdleConns = max
	return that
}

func (that *Builder) WithConnMaxLifetime(d time.Duration) *Builder {
	that.connMaxLifetime = d
	return that
}

func (that *Builder) WithConnMaxIdleTime(d time.Duration) *Builder {
	that.connMaxIdleTime = d
	return that
}

func (that *Builder) WithBeforeAcquire(fn func(ctx context.Context, conn *pgx.Conn) bool) *Builder {
	that.beforeAcquire = fn
	return that
}

func (that *Builder) WithQueryTracer(tracer pgx.QueryTracer) *Builder {
	that.tracer = tracer
	return that
}

func (that *Builder) Build() (DB, error) {
	ctx := context.Background()

	if err := that.checkRequiredParams(); err != nil {
		return nil, err
	}

	if err := that.updateDefaultParams(); err != nil {
		return nil, err
	}

	if err := that.errs.ResError(); err != nil {
		return nil, err
	}

	that.db.handler = that.newSniffer()

	that.db.source = that.dsn.String()
	config, err := pgxpool.ParseConfig(that.db.source)
	if err != nil {
		return nil, fmt.Errorf("ParseConfig: %w", err)
	}

	that.configureDB(config)

	config.ConnConfig.Tracer = that.tracer

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("Unable to create pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}

	that.db.pool = pool
	return &database{that.db}, nil
}

func (that *Builder) checkRequiredParams() error {
	if that.db.dbId == "" {
		return ErrDatabaseIDRequired
	}
	if that.dsn.Port == 0 {
		return ErrPortIsRequired
	}
	if that.dsn.Host == "" {
		return ErrHostIsRequired
	}
	if that.dsn.User == "" {
		return ErrUserIsRequired
	}
	if that.dsn.Database == "" {
		return ErrDatabaseIsRequired
	}
	return nil
}

func (that *Builder) updateDefaultParams() error {
	if that.db.errors == nil {
		that.db.errors = new(dummyErrorBuilder)
	}
	return nil
}

func (that *Builder) configureDB(config *pgxpool.Config) {
	config.ConnConfig.TLSConfig = nil

	if that.maxOpenConns > 0 {
		config.MaxConns = int32(that.maxOpenConns)
	}

	if that.maxIdleConns > 0 {
		config.MinConns = int32(that.maxIdleConns)
	}

	if that.connMaxLifetime > 0 {
		config.MaxConnLifetime = that.connMaxLifetime
	}

	if that.connMaxIdleTime > 0 {
		config.MaxConnIdleTime = that.connMaxIdleTime
	}

	config.BeforeAcquire = that.beforeAcquire
}

func (that *Builder) newSniffer() (handler Handler) {
	handler = NewDummyHandler()
	for i := len(that.middlewares) - 1; i >= 0; i-- {
		middleware := that.middlewares[i]
		handler = middleware(handler)
	}
	return handler
}

var (
	ErrDatabaseIDRequired = fmt.Errorf("DatabaseID is required")
	ErrPortIsRequired     = fmt.Errorf("Port is required")
	ErrHostIsRequired     = fmt.Errorf("Host is required")
	ErrUserIsRequired     = fmt.Errorf("User is required")
	ErrDatabaseIsRequired = fmt.Errorf("Database is required")
)
