package gsp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/airspace-link-inc/airhub-go/core/model"
	"github.com/stretchr/testify/assert"
	"github.com/uber/h3-go"
)

type asyncTestArgs struct {
	E       Pipeline
	Pipe    chan any
	ErrChan chan error
	Done    chan struct{}
}

// if this test fails CI/CD, then it may be the timeout values I chose.
// run locally and try to reproduce to see if this is the issue
const timeoutFirst = time.Millisecond * 500
const timeoutSecond = time.Millisecond * 600

var mockErr = fmt.Errorf("ErrAfter called")

func mockMetadataRows(resourceType ...string) []model.Metadata {
	m := make([]model.Metadata, len(resourceType))
	for i, v := range resourceType {
		m[i] = model.Metadata{
			ResourceType: v,
			Code:         "aiondoia",
			Resource:     "asndoa",
		}
	}
	return m
}

var spitsOut = struct {
	ValueAfter  func(t time.Duration) Synchronous
	ValuesAfter func(t time.Duration) Synchronous
	ErrAfter    func(t time.Duration) Synchronous
}{
	ValuesAfter: func(t time.Duration) Synchronous {
		return Synchronous{
			resourceType: "mocks",
			fn: func(ctx context.Context, s *semaphore, rows ...model.Metadata) ([]h3geo.D, error) {
				time.Sleep(t)

				features := make([]h3geo.D, len(rows))
				for i := range rows {
					features[i] = h3geo.HexFeature{Hexes: []h3.H3Index{}, Props: map[string]any{"something": i}}
				}

				return features, nil
			},
		}
	},
	ValueAfter: func(t time.Duration) Synchronous {
		return Synchronous{
			resourceType: "mock",
			fn: func(ctx context.Context, s *semaphore, rows ...model.Metadata) ([]h3geo.D, error) {
				time.Sleep(t)

				return []h3geo.D{
					h3geo.HexFeature{Hexes: []h3.H3Index{}, Props: map[string]any{"something": 12}},
				}, nil
			},
		}
	},
	ErrAfter: func(t time.Duration) Synchronous {
		return Synchronous{
			resourceType: "err",
			fn: func(ctx context.Context, s *semaphore, rows ...model.Metadata) ([]h3geo.D, error) {
				time.Sleep(t)
				return nil, mockErr
			},
		}
	},
}

func TestQuery(mainTest *testing.T) {
	testCases := []struct {
		name    string
		workers []Worker
		arg     []model.Metadata
		ctxGen  func() (context.Context, context.CancelFunc)
	}{
		{
			name: "base case",
		},
		{
			name: "properly handles a context that has already been canceled",
			arg: []model.Metadata{{
				Code:         "mock",
				ResourceType: "mock",
				Resource:     "iaonsd",
			}},
			workers: []Worker{
				spitsOut.ValueAfter(time.Millisecond * 200),
			},
			ctxGen: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
		},
		{
			name: "handles an error being thrown",
			arg: []model.Metadata{{
				Code:         "mock",
				ResourceType: "err",
				Resource:     "iaonsd",
			}},
			workers: []Worker{
				spitsOut.ErrAfter(0),
			},
		},
		{
			name: "handles an error and a value, giving preference to the error",
			arg: []model.Metadata{
				{
					Code:         "mock",
					ResourceType: "err",
					Resource:     "iaonsd",
				},
				{
					Code:         "mock",
					ResourceType: "mock",
					Resource:     "iaonsd",
				},
			},
			workers: []Worker{
				spitsOut.ErrAfter(100),
				spitsOut.ValueAfter(0),
			},
		},
		{
			name: "handles multiple h3geo.D return values",
			ctxGen: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			arg: []model.Metadata{
				{
					Code:         "mock1",
					ResourceType: "mocks",
					Resource:     "Mock resource 1",
				},
				{
					Code:         "mock2",
					ResourceType: "mocks",
					Resource:     "Mock resource 2",
				},
				{
					Code:         "mock3",
					ResourceType: "mocks",
					Resource:     "Mock resource 3",
				},
			},
			workers: []Worker{
				spitsOut.ValuesAfter(0),
			},
		},
	}

	config := util.GenCupaloyConfig()
	for i := range testCases {
		tc := testCases[i]
		if tc.ctxGen == nil {
			tc.ctxGen = defaultCtx
		}

		ctx, cancel := tc.ctxGen()
		defer cancel()

		previousBucketWorkers := genBucketWorkers
		defer func() {
			genBucketWorkers = previousBucketWorkers
		}()

		genBucketWorkers = func() []Worker { return tc.workers }
		actual, actualErr := Pipeline{}.Query(ctx, tc.arg...)
		mainTest.Run(tc.name, func(t *testing.T) {
			config.SnapshotT(t, actual, actualErr)
		})
	}
}

func TestAsyncQuery(mainTest *testing.T) {
	// single value tests
	testSingleValues(mainTest)

	// test multiple values
	previousBucketWorkers := genBucketWorkers
	defer func() {
		genBucketWorkers = previousBucketWorkers
	}()

	ctx, cancel := defaultCtx()
	defer cancel()

	genBucketWorkers = func() []Worker {
		return []Worker{
			spitsOut.ValueAfter(0),
			Synchronous{
				resourceType: "another worker",
				fn: func(ctx context.Context, s *semaphore, rows ...model.Metadata) ([]h3geo.D, error) {
					time.Sleep(time.Millisecond * 50)
					return []h3geo.D{
						h3geo.HexFeature{Hexes: []h3.H3Index{}, Props: map[string]any{"something else": 132}},
					}, nil
				},
			},
			Synchronous{
				resourceType: "3rd worker",
				fn: func(ctx context.Context, s *semaphore, rows ...model.Metadata) ([]h3geo.D, error) {
					time.Sleep(time.Millisecond * 100)
					return []h3geo.D{
						h3geo.HexFeature{Hexes: []h3.H3Index{}, Props: map[string]any{"aosidn": 12123}},
					}, nil
				},
			},
		}
	}

	args := newTestArgs()
	go runAsync(ctx, args, mockMetadataRows("mock", "another worker", "3rd worker")...)

	first, second, third := channelRead(args.Pipe), channelRead(args.Pipe), channelRead(args.Pipe)
	doneSignal, actualErr := channelRead(args.Done), channelRead(args.ErrChan)

	t := assert.New(mainTest)

	t.Equal(nil, actualErr, "With multiple workers that resolve, there should be no error")
	t.Equal(struct{}{}, doneSignal, "With multiple workers that resolve, there should a done signal")

	t.Equal(
		h3geo.HexFeature{
			Hexes: []h3.H3Index{},
			Props: map[string]any{"something": 12},
		},
		first,
		"Synchronizes the first return value from an async request",
	)
	t.Equal(
		h3geo.HexFeature{
			Hexes: []h3.H3Index{},
			Props: map[string]any{"something else": 132},
		},
		second,
		"Synchronizes the 2nd return value from an async request",
	)
	t.Equal(
		h3geo.HexFeature{
			Hexes: []h3.H3Index{},
			Props: map[string]any{"aosidn": 12123},
		},
		third,
		"Synchronizes the 3rd return value from an async request",
	)
}

func testSingleValues(mainTest *testing.T) {
	tests := []struct {
		name string
		fn   func() [3]any // element 0 is the h3geo.D val, 1 is error, 2 is done channel
	}{
		{
			"doesn't deadlock with no arguments",
			func() [3]any {
				ctx, cancel := defaultCtx()
				defer cancel()

				args := newTestArgs()
				go runAsync(ctx, args)

				// order matters!
				done, err, pipe := channelRead(args.Done), channelRead(args.ErrChan), channelRead(args.Pipe)
				return [3]any{pipe, err, done}
			},
		},
		{
			"doesn't deadlock with no arguments with a canceled context",
			func() [3]any {
				ctx, cancel := defaultCtx()
				defer cancel()

				args := newTestArgs()
				go runAsync(ctx, args)

				// order matters!
				done, err, pipe := channelRead(args.Done), channelRead(args.ErrChan), channelRead(args.Pipe)
				return [3]any{pipe, err, done}
			},
		},
		{
			"doesn't deadlock with a context that's been canceled when it wants to throw an err",
			func() [3]any {
				ctx, cancel := defaultCtx()
				cancel()

				args := newTestArgs()
				go runAsync(ctx, args, model.Metadata{})

				// order matters!
				done, err, pipe := channelRead(args.Done), channelRead(args.ErrChan), channelRead(args.Pipe)
				return [3]any{pipe, err, done}
			},
		},
		{
			"error preprocessing (missing code) is sent without deadlock",
			func() [3]any {
				ctx, cancel := defaultCtx()
				defer cancel()

				args := newTestArgs()
				go runAsync(ctx, args, model.Metadata{}) // is missing

				// order matters!
				return [3]any{
					channelRead(args.ErrChan),
					channelRead(args.Done),
					channelRead(args.Pipe),
				}
			},
		},
		{
			// if this test fails CI/CD, then it may be the timeout values I chose.
			// run locally and try to reproduce
			"syncs up one worker",
			func() [3]any {
				ctx, cancel := context.WithTimeout(context.Background(), timeoutFirst)
				defer cancel()

				args := newTestArgs()
				genBucketWorkers = func() []Worker {
					return []Worker{spitsOut.ValueAfter(0)}
				}

				go runAsync(ctx, args, model.Metadata{
					ResourceType: "mock",
					Code:         "oinaf",
					Resource:     "aiosnd",
				})

				// order matters!
				return [3]any{
					channelRead(args.Pipe),
					channelRead(args.Done),
					channelRead(args.ErrChan),
				}
			},
		},
		{
			// if this test fails CI/CD, then it may be the timeout values I chose.
			// run locally and try to reproduce
			"properly blocks sending a value if there's an error",
			func() [3]any {
				ctx, cancel := context.WithTimeout(context.Background(), timeoutFirst)
				defer cancel()

				args := newTestArgs()
				genBucketWorkers = func() []Worker {
					return []Worker{
						spitsOut.ValueAfter(time.Millisecond * 50),
						spitsOut.ErrAfter(0),
					}
				}

				go runAsync(
					ctx,
					args,
					mockMetadataRows("mock", "err")...,
				)

				// order matters!
				return [3]any{
					channelRead(args.ErrChan),
					channelRead(args.Done),
					channelRead(args.Pipe),
				}
			},
		},
		{
			// if this test fails CI/CD, then it may be the timeout values I chose.
			// run locally and try to reproduce
			"properly blocks sending several errors",
			func() [3]any {
				ctx, cancel := context.WithTimeout(context.Background(), timeoutFirst)
				defer cancel()

				args := newTestArgs()
				genBucketWorkers = func() []Worker {
					return []Worker{
						spitsOut.ErrAfter(0),
					}
				}

				go runAsync(
					ctx,
					args,
					mockMetadataRows("err", "err", "err")...,
				)

				// order matters!
				return [3]any{
					channelRead(args.ErrChan),
					channelRead(args.Done),
					channelRead(args.Pipe),
				}
			},
		},
	}

	snapshotter := util.GenCupaloyConfig()
	for i := range tests {
		tc := tests[i]
		previousBucketWorkers := genBucketWorkers
		defer func() {
			genBucketWorkers = previousBucketWorkers
		}()

		mainTest.Run(tc.name, func(t *testing.T) {
			values := tc.fn()
			snapshotter.SnapshotT(t,
				values[0],
				values[1],
				values[2],
			)
		})

		// reset bucket workers every iteration
		genBucketWorkers = previousBucketWorkers
	}
}

// generate all the args needed for the test. mostly pre-fills channels
func newTestArgs(partial ...asyncTestArgs) *asyncTestArgs {
	var full *asyncTestArgs
	switch len(partial) {
	case 0:
		full = &asyncTestArgs{}
	case 1:
		full = &partial[0]
	default:
		panic("only supply one to be ergonomic: newTestArgs() or newTestArgs(asyncTestArgs{/* some values filled */})")
	}

	if full.Pipe == nil {
		full.Pipe = make(chan h3geo.D)
	}

	if full.ErrChan == nil {
		full.ErrChan = make(chan error)
	}

	if full.Done == nil {
		full.Done = make(chan struct{})
	}

	return full
}

func defaultCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*2)
}

func runAsync(ctx context.Context, args *asyncTestArgs, rows ...model.Metadata) {
	args.E.AsyncQuery(ctx, args.Pipe, args.ErrChan, args.Done, rows...)
}

func channelRead[X any](c chan X) any {
	ctx, _ := defaultCtx()

	// either read or timeout
	select {
	case value := <-c:
		return value
	case <-ctx.Done():
		return ctx.Err()
	}
}
