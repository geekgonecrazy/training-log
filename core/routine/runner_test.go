package routine

import "testing"

func TestNextItem_Forward(t *testing.T) {
	items := []int64{10, 20, 30}
	states := []ItemState{StatePending, StatePending, StatePending}

	// Start: no cursor, go forward → first item.
	if i, done := NextItem(items, states, -1, true); done || i != 0 {
		t.Fatalf("got (%d, %v), want (0, false)", i, done)
	}
}

func TestNextItem_SkipLoopsBack(t *testing.T) {
	// User skipped item 0 and 2, completed item 1, now at end. Should loop back to 0.
	items := []int64{10, 20, 30}
	states := []ItemState{StateSkipped, StateCompleted, StateSkipped}

	i, done := NextItem(items, states, 2, true)
	if done {
		t.Fatal("expected not done")
	}
	if i != 0 {
		t.Fatalf("expected loop-back to 0, got %d", i)
	}
}

func TestNextItem_AllComplete(t *testing.T) {
	items := []int64{10, 20}
	states := []ItemState{StateCompleted, StateCompleted}
	if _, done := NextItem(items, states, 1, true); !done {
		t.Fatal("expected done")
	}
}

func TestNextItem_OnlyCompletedRemain(t *testing.T) {
	// Item 0 done, item 1 pending. From cursor=0, next should be 1.
	items := []int64{10, 20}
	states := []ItemState{StateCompleted, StatePending}
	i, done := NextItem(items, states, 0, true)
	if done || i != 1 {
		t.Fatalf("got (%d, %v), want (1, false)", i, done)
	}
}

func TestNextItem_Backward(t *testing.T) {
	items := []int64{10, 20, 30}
	states := []ItemState{StatePending, StateCompleted, StatePending}
	// At cursor=2, going backward should skip the completed 1 and land on 0.
	i, done := NextItem(items, states, 2, false)
	if done || i != 0 {
		t.Fatalf("got (%d, %v), want (0, false)", i, done)
	}
}
