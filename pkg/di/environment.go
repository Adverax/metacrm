package di

import "context"

type Environment struct {
	app Application
	ctx context.Context
}

func (that *Environment) Init() error {
	return that.app.Run(that.ctx)
}

func (that *Environment) Done() {
	that.app.Done(that.ctx)
}

func (that *Environment) Context() context.Context {
	return that.ctx
}

func NewEnvironment(
	ctx context.Context,
	constructor Constructor[Application],
	options ...AppOption,
) *Environment {
	opts := buildAppOptions(options...)

	app, ctx := build(ctx, constructor, opts)

	app.Setup(ctx)
	app.Init(ctx)

	return &Environment{app: app, ctx: ctx}
}
