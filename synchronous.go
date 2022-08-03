package gsp

import (
	"context"
	"sync"
)

type Synchronous[ArgType any] struct {
	B4 func(ArgType)
	Fn func(context.Context, chan<- Atom, chan<- error, ArgType) (Atom, error)
}

func (s *Synchronous[X]) Before(sem X) {
	if s.B4 != nil {
		s.B4(sem)
	}
}

func (s *Synchronous[X]) Fetch(ctx context.Context, pipe chan<- Atom, errChan chan<- error, wg *sync.WaitGroup, sem X) {
	defer wg.Done()

	resp, err := s.Fn(ctx, pipe, errChan, sem)
	if err != nil {
		select {
		case <-ctx.Done():
		case errChan <- err:
		}

		return
	}

	// nil is ignored in the event you just want to publish data to subscribers
	if resp == nil {
		return
	}

	select {
	case <-ctx.Done():
		return
	case pipe <- resp:
	}
}
