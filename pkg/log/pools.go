package log

import "sync"

type pool[T any] struct {
	pool *sync.Pool
}

func newPool[T any]() *pool[T] {
	return &pool[T]{
		pool: &sync.Pool{
			New: func() interface{} {
				return new(T)
			},
		},
	}
}

func (that *pool[T]) Put(entry *T) {
	that.pool.Put(entry)
}

func (that *pool[T]) Get() *T {
	return that.pool.Get().(*T)
}
