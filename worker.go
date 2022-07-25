package gsp

import (
	"context"
	"sync"

	"github.com/airspace-link-inc/airhub-go/core/h3geo"
	"github.com/airspace-link-inc/airhub-go/core/model"
)

type Worker[Out any, Argument sync.Locker] interface {
	// Before is used to subscribe to channel events, if needed. If you
	// need a token for example, this is where you subscribe to get it
	Before(s *semaphore)

	// Query is the code that resolves the metadata rows
	Go(ctx context.Context, pipe chan<- Out, errChan chan<- error, wg *sync.WaitGroup, args Argument)
}

// SynchronousBucket is the simplest implementation of a bucket worker that gets everything it
// needs synchronously
type Synchronous[X any] struct {
	Store X
	Fn    func(ctx context.Context, s *semaphore, rows ...model.Metadata) ([]h3geo.D, error)
	B4    func(store *X, s *semaphore)
}

func (s Synchronous[X]) Before(sem *semaphore) {
	s.B4(&s.Store, sem)
}

func (s Synchronous[X]) Go(ctx context.Context, pipe chan<- Out, errChan chan<- error, wg *sync.WaitGroup, args Argument) {
	defer wg.Done()

	if len(rows) == 0 {
		return
	}

	resps, err := s.fn(ctx, sem, rows...)
	if err != nil {
		select {
		case <-ctx.Done():
		case sem.errChan <- err:
		}
		return
	}

	for _, v := range resps {
		select {
		case <-ctx.Done():
			return
		case sem.pipe <- v:
		}
	}
}
