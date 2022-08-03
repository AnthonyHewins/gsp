package gsp

import (
	"context"
	"sync"
)

// SynchronousWithStore is an implementation of the insight interface
// in a SynchronousWithStore fashion, where the implementation returns
// all its data at once
type SynchronousWithStore[ArgType any, Store any] struct {
	Store Store
	B4    func(ArgType, *SynchronousWithStore[ArgType, Store])
	Fn    func(context.Context, chan<- Atom, chan<- error, ArgType, *SynchronousWithStore[ArgType, Store]) (Atom, error)
}

func (s *SynchronousWithStore[ArgType, X]) Before(sem ArgType) {
	if s.B4 != nil {
		s.B4(sem, s)
	}
}

func (s *SynchronousWithStore[ArgType, X]) Fetch(ctx context.Context, pipe chan<- Atom, errChan chan<- error, wg *sync.WaitGroup, sem ArgType) {
	defer wg.Done()

	resp, err := s.Fn(ctx, pipe, errChan, sem, s)
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
