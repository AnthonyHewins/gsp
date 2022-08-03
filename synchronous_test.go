package gsp

import "testing"

func TestSyncWithStore(t *testing.T) {
	correct := 101293

	mockStore := SynchronousWithStore[struct{}, int]{
		B4: func(s struct{}, sws *int) {
			*sws = correct
		},
	}

	mockStore.B4(struct{}{}, &mockStore.Store)
	if mockStore.Store != correct {
		t.Errorf("b4 should have set synchronousWithStore.store to %v, got %v", correct, mockStore.Store)
	}
}
