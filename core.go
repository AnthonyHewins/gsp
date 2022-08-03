package gsp

import (
	"reflect"
	"sync"

	"context"
)

func Async[ArgType any](ctx context.Context, pipe chan<- Atom, errChan chan<- error, done chan<- struct{}, arg ArgType, jobs ...Job[ArgType]) {
	defer close(errChan)
	defer close(pipe)
	defer close(done)
	defer func() { done <- struct{}{} }()

	n := len(jobs)
	if n == 0 {
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(n)
	for _, job := range jobs {
		go job.Fetch(ctx, pipe, errChan, &wg, arg)
	}

	wg.Wait()
}

func New[X Value[Arg], Arg any](ctx context.Context, sem Arg) (*X, error) {
	var value X
	done, pipe, errChan := make(chan struct{}), make(chan Atom), make(chan error)
	go Async(ctx, pipe, errChan, done, sem, value.Jobs()...)

	t := reflect.TypeOf(value)
	v := reflect.ValueOf(&value).Elem()

	for {
		select {
		case <-ctx.Done():
			return &value, ctx.Err()
		case <-done:
			return &value, nil
		case err := <-errChan:
			return &value, err
		case field := <-pipe:
			field.Assign(t, v)
		}
	}
}
