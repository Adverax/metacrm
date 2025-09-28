package di

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Define components
type MyEvents struct {
	// Your events here
}

func newMyEvents() *MyEvents {
	return &MyEvents{}
}

type MyRepository struct {
	filename string
	events   *MyEvents
}

type MyRepositoryBuilder struct {
	repo *MyRepository
}

func NewMyRepositoryBuilder() *MyRepositoryBuilder {
	return &MyRepositoryBuilder{
		repo: &MyRepository{},
	}
}

func (b *MyRepositoryBuilder) WithFilename(filename string) *MyRepositoryBuilder {
	b.repo.filename = filename
	return b
}

func (b *MyRepositoryBuilder) WithEvents(events *MyEvents) *MyRepositoryBuilder {
	b.repo.events = events
	return b
}

func (b *MyRepositoryBuilder) Build() (*MyRepository, error) {
	if err := b.checkRequiredFields(); err != nil {
		return nil, err
	}

	return b.repo, nil
}

func (b *MyRepositoryBuilder) checkRequiredFields() error {
	if b.repo.filename == "" {
		return fmt.Errorf("filename is required")
	}

	if b.repo.events == nil {
		return fmt.Errorf("events is required")
	}

	return nil
}

type MyScheduler struct {
	events *MyEvents
}

func (s *MyScheduler) Start() error {
	fmt.Println("Scheduler started")
	return nil
}

type MySchedulerBuilder struct {
	scheduler *MyScheduler
}

func newMySchedulerBuilder() *MySchedulerBuilder {
	return &MySchedulerBuilder{
		scheduler: &MyScheduler{},
	}
}

func (b *MySchedulerBuilder) WithEvents(events *MyEvents) *MySchedulerBuilder {
	b.scheduler.events = events
	return b
}

func (b *MySchedulerBuilder) Build() (*MyScheduler, error) {
	if err := b.checkRequiredFields(); err != nil {
		return nil, err
	}

	return b.scheduler, nil
}

func (b *MySchedulerBuilder) checkRequiredFields() error {
	if b.scheduler.events == nil {
		return fmt.Errorf("events is required")
	}

	return nil
}

type MyApplication struct {
	*App
	Events     *MyEvents
	Repository *MyRepository
}

type MyApplicationBuilder struct {
	app *MyApplication
}

func NewMyApplicationBuilder() *MyApplicationBuilder {
	return &MyApplicationBuilder{
		app: &MyApplication{},
	}
}

func (b *MyApplicationBuilder) WithApp(app *App) *MyApplicationBuilder {
	b.app.App = app
	return b
}

func (b *MyApplicationBuilder) WithEvents(events *MyEvents) *MyApplicationBuilder {
	b.app.Events = events
	return b
}

func (b *MyApplicationBuilder) WithRepository(repo *MyRepository) *MyApplicationBuilder {
	b.app.Repository = repo
	return b
}

func (b *MyApplicationBuilder) Build() (*MyApplication, error) {
	if err := b.checkRequiredFields(); err != nil {
		return nil, err
	}

	return b.app, nil
}

func (b *MyApplicationBuilder) checkRequiredFields() error {
	if b.app.App == nil {
		return fmt.Errorf("app is required")
	}

	if b.app.Events == nil {
		return fmt.Errorf("events is required")
	}

	if b.app.Repository == nil {
		return fmt.Errorf("repository is required")
	}

	return nil
}

var ComponentEnvironment = NewComponent(
	"MyEnvironment",
	func(ctx context.Context) (*MyEnvironment, error) {
		return GetEnvironmentFromContext(ctx), nil
	},
)

// Declare components
var ComponentApplication = NewComponent(
	"MyApplication",
	func(ctx context.Context) (Application, error) {
		return NewMyApplicationBuilder().
			WithRepository(ComponentRepository(ctx)).
			WithEvents(ComponentEvents(ctx)).
			WithApp(GetAppFromContext(ctx)).
			Build()
	},
)

var ComponentEvents = NewComponent(
	"MyEvents",
	func(ctx context.Context) (*MyEvents, error) {
		return newMyEvents(), nil
	},
)

var ComponentRepository = NewComponent(
	"MyRepository",
	func(ctx context.Context) (*MyRepository, error) {
		env := ComponentEnvironment(ctx)
		return NewMyRepositoryBuilder().
			WithFilename(env.Config.RepositoryFileName).
			WithEvents(ComponentEvents(ctx)).
			Build()
	},
)

var ComponentScheduler = NewComponent(
	"Scheduler",
	func(ctx context.Context) (*MyScheduler, error) {
		return newMySchedulerBuilder().
			WithEvents(ComponentEvents(ctx)).
			Build()
	},
	WithComponentInit(func(ctx context.Context, instance *MyScheduler) error {
		return instance.Start()
	}),
)

type MyConfig struct {
	// Your config here
	RepositoryFileName string
}

type MyEnvironment struct {
	// Your environment here
	Config *MyConfig
}

type EnvironmentContextType int

var EnvironmentContextKey EnvironmentContextType = 0

func GetEnvironmentFromContext(ctx context.Context) *MyEnvironment {
	return ctx.Value(EnvironmentContextKey).(*MyEnvironment)
}

// Example of dependency injection usage
func Example() {
	// Handle signals for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-sigs
		cancel()
	}()

	// Allow 5 seconds for lifecycle of our application
	ctx, cancel2 := context.WithTimeout(ctx, 5*time.Second)
	defer cancel2()

	// Define environment
	env := &MyEnvironment{
		Config: &MyConfig{
			RepositoryFileName: "data.json",
		},
	}

	ctx = context.WithValue(ctx, EnvironmentContextKey, env)

	// Execute application
	Execute(ctx, ComponentApplication)

	// Output:
	// Scheduler started
}
