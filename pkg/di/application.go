package di

import (
	"context"
	"fmt"
	"sync"

	"github.com/adverax/metacrm/pkg/access"
	"github.com/adverax/metacrm/pkg/containers/maps"
)

type Variables = access.ReaderWriter

type Daemon = constructor

type constructor func(ctx context.Context)

type components []*component

// Application - interface for application
type Application interface {
	// Daemons - must return list of daemons
	Daemons(ctx context.Context) []Daemon
	// Setup - called to setup application
	Setup(ctx context.Context)
	// Init - called to initialize application
	Init(ctx context.Context)
	// Run - called to run application
	Run(ctx context.Context) error
	// Done - called to finalize application
	Done(ctx context.Context)
}

func newApp(options AppOptions) *App {
	return &App{
		dictionary: make(map[string]*component),
		variables:  make(maps.Map),
		logger:     options.logger,
	}
}

// App - base implementation of Application
type App struct {
	mx         sync.Mutex
	components components
	dictionary map[string]*component
	variables  Variables
	logger     Logger
}

func (that *App) Setup(_ context.Context) {
	// empty
}

func (that *App) Daemons(_ context.Context) []Daemon {
	return nil
}

func (that *App) Variables() Variables {
	return that.variables
}

func (that *App) addComponent(component *component) {
	that.mx.Lock()
	defer that.mx.Unlock()

	that.components = append(that.components, component)
	that.dictionary[component.name] = component
}

func (that *App) Init(ctx context.Context) {
	for _, c := range that.components {
		if err := c.runInit(ctx, that.logger); err != nil {
			panic(&componentError{c.name, fmt.Sprintf("init: %s", err.Error())})
		}
	}
}

func (that *App) Done(ctx context.Context) {
	cs := that.components
	for i := len(cs) - 1; i >= 0; i-- {
		c := cs[i]
		c.runDone(ctx, that.logger)
	}
}

func (that *App) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (that *App) get(ctx context.Context, name string, builder func(ctx context.Context, app *App) *component) *component {
	c := that.fetch(ctx, name)
	if c != nil {
		return c
	}

	c = builder(ctx, that)
	that.addComponent(c)

	return c
}

func (that *App) fetch(_ context.Context, name string) *component {
	that.mx.Lock()
	defer that.mx.Unlock()

	c, _ := that.dictionary[name]
	return c
}

func (that *App) set(component *component) {
	that.mx.Lock()
	defer that.mx.Unlock()

	if _, exists := that.dictionary[component.name]; exists {
		panic(fmt.Errorf("service %s already exists", component.name))
	}

	that.components = append(that.components, component)
	that.dictionary[component.name] = component
}

type AppOptions struct {
	logger Logger
}

type AppOption func(opts *AppOptions)

// WithAppLogger - set logger for application
func WithAppLogger(logger Logger) AppOption {
	return func(opts *AppOptions) {
		opts.logger = logger
	}
}

// Execute - primary entry point for build and run application
func Execute(
	ctx context.Context,
	constructor Constructor[Application],
	options ...AppOption,
) error {
	opts := buildAppOptions(options...)

	app, ctx := build(ctx, constructor, opts)

	app.Setup(ctx)
	app.Init(ctx)
	defer app.Done(ctx)

	return app.Run(ctx)
}

// Build - build application
func Build(
	ctx context.Context,
	constructor Constructor[Application],
	options ...AppOption,
) (a Application, c context.Context) {
	opts := buildAppOptions(options...)
	return build(ctx, constructor, opts)
}

func build(
	ctx context.Context,
	constructor Constructor[Application],
	opts AppOptions,
) (a Application, c context.Context) {
	app := newApp(opts)
	ctx = context.WithValue(ctx, ApplicationContextKey, app)

	application := constructor(ctx)

	for _, c := range application.Daemons(ctx) {
		c(ctx)
	}

	return application, ctx
}

type ApplicationContextType int

var ApplicationContextKey ApplicationContextType = 0

// GetAppFromContext - get application from context
func GetAppFromContext(ctx context.Context) *App {
	app, _ := ctx.Value(ApplicationContextKey).(*App)
	if app == nil {
		panic(fmt.Errorf("application not found in context"))
	}
	return app
}

// SetAppToContext - set application to context
func SetAppToContext(ctx context.Context, app *App) context.Context {
	return context.WithValue(ctx, ApplicationContextKey, app)
}

// GetVariable - get variable from application context
func GetVariable[T any](ctx context.Context, key string) (val T, err error) {
	app := GetAppFromContext(ctx)
	vars := app.Variables()
	return access.GetValue[T](ctx, vars, key)
}

// SetVariable - set variable to application context
func SetVariable[T any](ctx context.Context, key string, val T) error {
	app := GetAppFromContext(ctx)
	vars := app.Variables()
	return vars.SetProperty(ctx, key, val)
}

func buildAppOptions(options ...AppOption) AppOptions {
	opts := AppOptions{
		logger: &dummyLogger{},
	}
	for _, o := range options {
		o(&opts)
	}
	return opts
}
