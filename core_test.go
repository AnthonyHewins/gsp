package gsp

import (
	"context"
	"fmt"
	"testing"
)

var mockJobs []Job[int]

type mockInsight struct {
	F int
	J int
}

func (m mockInsight) Jobs() []Job[int] {
	return mockJobs
}

func TestInsight(test *testing.T) {
	returnGJForField := func(i int) *Synchronous[int] {
		return &Synchronous[int]{
			Fn: func(ctx context.Context, pipe chan<- Atom, errChan chan<- error, s int) (Atom, error) {
				return &IndexAtom{Field: i, Data: 11}, nil
			},
		}
	}

	returnErr := &Synchronous[int]{
		Fn: func(ctx context.Context, pipe chan<- Atom, errChan chan<- error, s int) (Atom, error) {
			return nil, fmt.Errorf("error")
		},
	}

	tests := []struct {
		name string
		jobs []Job[int]
	}{
		{
			name: "base case",
		},
		{
			name: "handles one job returning a single feature",
			jobs: []Job[int]{returnGJForField(0)},
		},
		{
			name: "handles one job returning an error",
			jobs: []Job[int]{returnErr},
		},
		{
			name: "handles 2 jobs, both returning features",
			jobs: []Job[int]{returnGJForField(0), returnGJForField(1)},
		},
		{
			name: "handles 2 jobs, one producing an err, and returns the err only",
			jobs: []Job[int]{returnGJForField(0), returnErr},
		},
		{
			name: "handles several errored jobs, capturing errors correctly and not blowing up",
			jobs: []Job[int]{returnErr, returnErr, returnErr, returnErr, returnErr, returnErr},
		},
	}

	snapshotT := snapshotTest().SnapshotT
	for _, tc := range tests {
		test.Run(tc.name, func(t *testing.T) {
			mockJobs = tc.jobs
			actual, actualErr := New[mockInsight, int](context.Background(), 0)
			snapshotT(t, actual, actualErr)
		})
	}
}
