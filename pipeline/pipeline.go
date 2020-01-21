package pipeline

import (
	"context"
	"sync"
)

type Upstream interface {
	Wait() error
	Ctx() context.Context
	Run(f func() error)
	Done() <-chan struct{}
	New() Downstream
}

type Downstream interface {
	Upstream

	Stop()
}

type downstreamImpl struct {
	cancel func()

	wg sync.WaitGroup

	errOnce sync.Once
	err     error
	ctx     context.Context
}

func New(ctx context.Context) Downstream {
	ctx, cancel := context.WithCancel(ctx)

	return &downstreamImpl{cancel: cancel, ctx: ctx}
}

func (g *downstreamImpl) Wait() error {
	g.wg.Wait()
	g.errOnce.Do(func() {
		g.cancel()
	})
	
	return g.err
}

func (g *downstreamImpl) Stop() {
	g.errOnce.Do(func() {
		g.cancel()
	})
}

func (g *downstreamImpl) Ctx() context.Context {
	return g.ctx
}

func (g *downstreamImpl) Run(f func() error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				g.cancel()
			})
		}
	}()
}

func (g *downstreamImpl) Done() <-chan struct{} {
	return g.ctx.Done()
}

func (g *downstreamImpl) New() Downstream {
	ctx, cancel := context.WithCancel(g.ctx)

	downstream := &downstreamImpl{
		cancel:  cancel,
		ctx:     ctx,
	}

	g.Run(func() error {
		return downstream.Wait()
	})

	return downstream
}
