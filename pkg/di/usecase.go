package di

import (
	"context"
	"fmt"
)

const ConfigKey = "config"

type Action func(ctx context.Context) error

type usecase[T any] struct {
	*App
	config T
	action Action
}

func NewUsecase[T any](config T, action Action) Constructor[Application] {
	return func(ctx context.Context) Application {
		app := GetAppFromContext(ctx)
		err := SetVariable(ctx, ConfigKey, config)
		if err != nil {
			panic("failed to set config variable: " + err.Error())
		}
		return &usecase[T]{
			App:    app,
			config: config,
			action: action,
		}
	}
}

func (that *usecase[T]) Run(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("usecase panic: %v", r)
		}
	}()

	err = that.action(ctx)
	if err != nil {
		return err
	}

	that.App.Init(ctx)
	return nil
}
