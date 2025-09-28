package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adverax/metacrm.kernel/database/sql"
	"github.com/adverax/metacrm.kernel/di"
	"github.com/adverax/metacrm.kernel/log"
	fileExporter "github.com/adverax/metacrm.kernel/log/exporters/file"
	jsonFormatter "github.com/adverax/metacrm.kernel/log/formatters/json"
	templateFormatter "github.com/adverax/metacrm.kernel/log/formatters/template"
	"github.com/adverax/metacrm.kernel/log/purifiers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

var (
	ComponentConfig = di.NewComponent(
		"config",
		func(ctx context.Context) (*Config, error) {
			return di.GetVariable[*Config](ctx, di.ConfigKey)
		},
	)

	ComponentLogFormatter = di.NewComponent(
		"log-formatter",
		func(ctx context.Context) (log.Formatter, error) {
			cfg := ComponentConfig(ctx)
			switch cfg.Log.Format {
			case "text":
				return templateFormatter.NewBuilder().
					WithPurifier(purifiers.NewMultilinePurifier(nil)).
					Build()
			case "json":
				return jsonFormatter.NewBuilder().
					Build()
			default:
				return nil, fmt.Errorf("Unknown log format: %s", cfg.Log.Format)
			}
		},
	)

	ComponentSystemLogFile = di.NewComponent(
		"system-log-file",
		func(ctx context.Context) (*os.File, error) {
			filename := "var/log/dev.log"

			if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
				return nil, fmt.Errorf("failed to create log directory: %w", err)
			}

			_ = os.Remove(filename)
			f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, fmt.Errorf("failed to open system log file: %w", err)
			}
			return f, nil
		},
		di.WithComponentDone(func(ctx context.Context, instance *os.File) {
			instance.Close()
		}),
	)

	ComponentLogFile = di.NewComponent(
		"log-file",
		func(ctx context.Context) (*os.File, error) {
			cfg := ComponentConfig(ctx)
			if isDevEnv() {
				return ComponentSystemLogFile(ctx), nil
			}

			switch cfg.Log.Output {
			case "stdout":
				return os.Stdout, nil
			case "stderr":
				return os.Stderr, nil
			default:
				return nil, fmt.Errorf("Unknown log output: %s", cfg.Log.Output)
			}
		},
	)

	ComponentLogExporter = di.NewComponent(
		"log-exporter",
		func(ctx context.Context) (log.Exporter, error) {
			return fileExporter.New(
				ComponentLogFile(ctx),
				ComponentLogFormatter(ctx),
			), nil
		},
	)

	ComponentSecrets = di.NewComponent(
		"secrets",
		func(ctx context.Context) (map[string]log.Masker, error) {
			return map[string]log.Masker{
				"password":      nil,
				"token":         nil,
				"authorization": nil,
				"cookie":        nil,
				"secret":        nil,
				"card":          log.MaskIDCard,
				"cvv":           nil,
				"pin":           nil,
				"email":         log.MaskEmail,
				"phone":         log.MaskPhone,
				"ssn":           nil,
				"passport":      nil,
				"address":       nil,
				"refresh_token": nil,
				"private_key":   nil,
			}, nil
		},
	)

	ComponentLogExporterWithSecretPurifier = di.NewComponent(
		"log-exporter",
		func(ctx context.Context) (log.Exporter, error) {
			return log.NewSecurityExporter(
				ComponentSecrets(ctx),
				ComponentLogExporter(ctx),
			), nil
		},
	)

	ComponentLogger = di.NewComponent(
		"logger",
		func(ctx context.Context) (log.Logger, error) {
			cfg := ComponentConfig(ctx)
			return log.NewBuilder().
				WithLevel(logLevels.EncodeOrDefault(cfg.Log.Level, log.InfoLevel)).
				WithExporter(ComponentLogExporterWithSecretPurifier(ctx)).
				Build()
		},
	)

	ComponentDatabaseErrorBuilder = di.NewComponent(
		"database-error-builder",
		func(ctx context.Context) (sql.ErrorBuilder, error) {
			return sql.NewDatabaseErrorBuilder(), nil
		},
	)

	ComponentDatabaseQueryLogger = di.NewComponent(
		"database-query-logger",
		func(ctx context.Context) (pgx.QueryTracer, error) {
			return sql.NewLogger(
				ComponentLogger(ctx),
				sql.LogOptions{
					MaxTemplateLen:     500,
					MaxAllowedDuration: 200 * time.Millisecond,
				},
			), nil
		},
	)

	ComponentDatabase = di.NewComponent(
		sql.DatabaseComponentName,
		func(ctx context.Context) (sql.DB, error) {
			cfg := ComponentConfig(ctx)
			return sql.NewBuilder().
				WithHost(cfg.DB.Host).
				WithPort(cfg.DB.Port).
				WithUser(cfg.DB.User).
				WithPassword(cfg.DB.Password).
				WithDatabase(cfg.DB.Database).
				WithErrorBuilder(ComponentDatabaseErrorBuilder(ctx)).
				WithQueryTracer(ComponentDatabaseQueryLogger(ctx)).
				Build()
		},
		di.WithComponentDone(func(ctx context.Context, instance sql.DB) {
			instance.Close()
		}),
	)

	ComponentRouter = di.NewComponent(
		"router",
		func(ctx context.Context) (*gin.Engine, error) {
			return gin.Default(), nil
		},
	)
)
