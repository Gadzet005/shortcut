package lifecycle

import (
	"context"
	"sync"
)

type Context interface {
	Context() context.Context
	AddStopper(stopper Stopper)
	// Запуск задачи, которая захватывает поток выполнения
	RunJob(runner Runner, stopper Stopper)
}

func NewContext(ctx context.Context) *contextImpl {
	return &contextImpl{
		ctx: ctx,
	}
}

type contextImpl struct {
	ctx      context.Context
	stoppers []Stopper
	wg       sync.WaitGroup
}

func (c *contextImpl) Context() context.Context {
	return c.ctx
}

func (c *contextImpl) AddStopper(stopper Stopper) {
	c.wg.Add(1)
	c.stoppers = append(c.stoppers, func(ctx context.Context) error {
		c.wg.Done()
		return stopper(ctx)
	})
}

func (c *contextImpl) RunJob(runner Runner, stopper Stopper) {
	c.stoppers = append(c.stoppers, stopper)
	c.wg.Go(func() {
		runner(c.ctx)
	})
}
