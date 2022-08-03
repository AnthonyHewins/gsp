package gsp

import (
	"context"
	"sync"
)

type Job[ArgType any] interface {
	Before(ArgType)
	Fetch(ctx context.Context, pipe chan<- Atom, errChan chan<- error, wg *sync.WaitGroup, sem ArgType)
}
