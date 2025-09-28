package log

import (
	"context"
	"sync"
)

type Hook interface {
	Fire(ctx context.Context, entry *Entry) error
}

type HookFunc func(ctx context.Context, entry *Entry) error

func (fn HookFunc) Fire(ctx context.Context, entry *Entry) error {
	return fn(ctx, entry)
}

type Hooks struct {
	sync.RWMutex
	hooks map[Level][]Hook
}

func NewHooks() *Hooks {
	return &Hooks{
		hooks: make(map[Level][]Hook),
	}
}

func (that *Hooks) Add(levels []Level, hook Hook) {
	that.Lock()
	defer that.Unlock()

	for _, level := range levels {
		that.hooks[level] = append(that.hooks[level], hook)
	}
}

func (that *Hooks) Fire(ctx context.Context, level Level, entry *Entry) error {
	that.RLock()
	defer that.RUnlock()

	for _, hook := range that.hooks[level] {
		if err := hook.Fire(ctx, entry); err != nil {
			return err
		}
	}

	return nil
}
