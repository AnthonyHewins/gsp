package gsp

import (
	"context"
	"sync"
)

type Pipeline[In any, Out any, PipelineArgs sync.Locker] struct {
	Workers []Worker[Out, PipelineArgs]
	Store   PipelineArgs
}

func (e Pipeline[In, Out, X]) Async(ctx context.Context, pipe chan<- Out, errChan chan<- error, done chan<- struct{}, input <-chan In) {
	defer close(errChan) // Order matters!
	defer close(pipe)
	defer close(done)
	defer func() { done <- struct{}{} }()

	if len(input) == 0 {
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s := semaphore{
		pipe:    pipe,
		errChan: errChan,
		wg:      sync.WaitGroup{},
	}

	s.wg.Add(len(e.Workers))
	for i, worker := range e.Workers {
		go bucket.Query(ctx, &s, sortedRows[i]...)
	}

	s.wg.Wait()
}

// Query is the synchronous way to have the metadata engine resolve data.
// Under the hood, it uses AsyncQuery.
// Upon getting an error, the entire routine will cancel and return.
// If the context is canceled or times out, it will return the error
func (e Pipeline[In, Out, PipelineArgs]) Query(ctx context.Context, rows ...In) ([]Out, error) {
	if len(rows) == 0 {
		return []Out{}, nil
	}

	var data []Out
	pipe, errChan, done := make(chan Out), make(chan error), make(chan struct{})
	go e.Async(ctx, pipe, errChan, done, rows...)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-done:
			return data, nil
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
		case feature := <-pipe:
			data = append(data, feature)
		}
	}
}
