package access

import (
	"context"
	"sync"
)

type SafeGetterSetter struct {
	sync.RWMutex
	GetterSetter
}

func (that *SafeGetterSetter) GetProperty(
	ctx context.Context,
	name string,
) (interface{}, error) {
	that.RLock()
	defer that.RUnlock()

	return that.GetterSetter.GetProperty(ctx, name)
}

func (that *SafeGetterSetter) SetProperty(
	ctx context.Context,
	name string,
	value interface{},
) error {
	that.Lock()
	defer that.Unlock()

	return that.GetterSetter.SetProperty(ctx, name, value)
}

func (that *SafeGetterSetter) Transaction(
	ctx context.Context,
	action func(ctx context.Context, gs GetterSetter) error,
) error {
	that.Lock()
	defer that.Unlock()

	return action(ctx, that.GetterSetter)
}

// NewSafeGetterSetter is constructor for build safe GetterSetter
func NewSafeGetterSetter(gs GetterSetter) *SafeGetterSetter {
	return &SafeGetterSetter{
		GetterSetter: gs,
	}
}
