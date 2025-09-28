package sql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/adverax/metacrm/pkg/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type LogOptions struct {
	MaxTemplateLen     int           // например, 500
	MaxAllowedDuration time.Duration // например, 200ms
}

func SafeTemplate(tmpl string, max int) string {
	t := strings.TrimSpace(tmpl)
	if len(t) > max {
		t = t[:max] + "…"
	}
	return t
}

type traceQueryStartData struct {
	*pgx.TraceQueryStartData
	started time.Time
}

type traceQueryStartKeyType int

var ctxKeyStart traceQueryStartKeyType = 1

type Tracer struct {
	options LogOptions
	logger  log.Logger
}

func NewLogger(logger log.Logger, options LogOptions) *Tracer {
	if options.MaxTemplateLen == 0 {
		options.MaxTemplateLen = 500
	}
	if options.MaxAllowedDuration == 0 {
		options.MaxAllowedDuration = 200 * time.Millisecond
	}

	return &Tracer{
		options: options,
		logger:  logger,
	}
}

func (that *Tracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	start := traceQueryStartData{started: time.Now(), TraceQueryStartData: &data}
	ctx = context.WithValue(ctx, ctxKeyStart, &start)
	return ctx
}

func (that *Tracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	start, _ := ctx.Value(ctxKeyStart).(*traceQueryStartData)
	rows := int(data.CommandTag.RowsAffected())
	that.log(
		ctx,
		strings.ToLower(data.CommandTag.String()),
		fmt.Sprintf("%d", conn.PgConn().PID()),
		start.SQL,
		start.Args,
		start.started,
		rows,
		data.Err,
	)
}

func (that *Tracer) log(
	ctx context.Context,
	op, resource, tmpl string,
	binds []any,
	start time.Time,
	rows int,
	err error,
) {
	d := time.Since(start)

	fields := log.Fields{
		"db.system":      "postgres",
		"db.operation":   op,
		"db.resource":    resource,
		"db.latency":     d,
		"db.rows":        rows,
		"db.fingerprint": SQLFingerprint(tmpl),
	}

	if that.options.MaxTemplateLen != 0 {
		fs := log.Fields{
			"sql": SafeTemplate(tmpl, that.options.MaxTemplateLen),
		}
		if len(binds) > 0 && that.logger.IsLevelEnabled(log.DebugLevel) {
			fs["binds"] = binds
		}
		fields["db.query"] = fs
	}

	logger := that.logger.WithFields(fields)

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		logger.Warning(ctx, "db_timeout", "sql.state", "57014") // cancellation timeout
	case err != nil:
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			logger = logger.WithField(
				"state",
				log.Fields{
					"code":       pgErr.Code,
					"table":      pgErr.TableName,
					"constraint": pgErr.ConstraintName,
				},
			)
		}
		logger.Error(ctx, "db_error")
	case d > that.options.MaxAllowedDuration:
		logger.Warning(ctx, "db_slow_query")
	default:
		logger.Debug(ctx, "db_query")
	}
}
