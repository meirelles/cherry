package postpone

import "sync"

type Postpone struct {
    queued []func() error
    isDone bool
    lock   sync.Mutex
}

func (p *Postpone) Run(fn func() error) {
    p.lock.Lock()
    defer p.lock.Unlock()

    if p.isDone {
        err := fn()

        if err != nil {
            panic(err)
        }
    } else {
        p.queued = append(p.queued, fn)
    }
}

func (p *Postpone) Done() {
    p.lock.Lock()
    defer p.lock.Unlock()

    for _, fn := range p.queued {
        err := fn()

        if err != nil {
            panic(err)
        }
    }

    p.isDone = true
}
