package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/adverax/metacrm/apps/backend/iam/bootstrap"
	"github.com/adverax/metacrm/pkg/di"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/stdlib"
)

type App struct {
	config *bootstrap.Config
}

func New() (*App, error) {
	config := bootstrap.DefaultConfig()

	err := config.Load()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not load config: %v", err))
	}

	err = config.Init()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not init config: %v", err))
	}

	err = config.Validate()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not validate config: %v", err))
	}

	return &App{config: config}, nil
}

func (that *App) StartServer() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return di.Execute(ctx, di.NewUsecase(that.config, that.execServe))
}

func (that *App) RunMigrations() error {
	return di.Execute(context.Background(), di.NewUsecase(that.config, that.execMigrations))
}

func (that *App) execServe(ctx context.Context) error {
	if that.config.IsDevEnv() {
		err := that.execMigrations(ctx)
		if err != nil {
			return err
		}
	}

	port := that.config.Api.Port

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: bootstrap.ComponentRouter(ctx),
	}

	serverErrCh := make(chan error)

	go func() {
		log.Printf("server is running... port=%d", port)
		defer log.Print("server gracefully stopped")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
	}()

	select {
	case err := <-serverErrCh:
		return errors.New(fmt.Sprintf("error starting server: %v", err))
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		log.Print("server is shutting down...")
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			err = errors.New(fmt.Sprintf("failed to shutdown server: %v", err))
		}
		return err
	}
}

func (that *App) execMigrations(ctx context.Context) error {
	db := bootstrap.ComponentDatabase(ctx)
	sqlDB := stdlib.OpenDBFromPool(db.Pool())

	driver, err := pgx.WithInstance(sqlDB, &pgx.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}
	defer func() {
		_ = driver.Close()
	}()

	m, err := migrate.NewWithDatabaseInstance("file://database/migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
