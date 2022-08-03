package gsp

import "testing"

func TestSyncWithStore(t *testing.T) {
	mockStore := SynchronousWithStore[struct{}, int]{
		B4: func(s struct{}, sws *SynchronousWithStore[struct{}, int]) {
			sws.Store = 1488
		},
	}

	mockStore.B4(struct{}{}, &mockStore)
	if mockStore.Store != 1488 {
		t.Errorf("b4 should have set synchronousWithStore.store to 1488, got %v", mockStore.Store)
	}
}
