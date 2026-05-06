// Package routine encodes the routine-runner state machine.
//
// The runner walks the routine's ordered items. Pressing "next" without
// completing logs a skip; at the end of the list the runner cycles back through
// any skipped items until they're all completed (or the user hits End). This
// matches the "drive me through everything" UX.
//
// The frontend implements the same logic locally; this Go version exists for
// server-side reasoning, tests, and any future API that needs to compute the
// "next" item authoritatively.
package routine

// ItemState records what's happened to an item during a run.
type ItemState int

const (
	StatePending   ItemState = iota // not yet acted on, in this pass
	StateCompleted                  // done in this run
	StateSkipped                    // skipped in this pass; eligible to retry
)

// NextItem chooses the index of the next item to present.
//
// items is the routine's ordered exercise IDs. states[i] is the corresponding
// ItemState for items[i]. cursor is the most recently presented index, or -1
// if nothing has been shown yet. forward=true means user pressed Next.
//
// Returns the chosen index and whether the routine is finished (no items left
// to visit). When finished, the second value is true and the first is undefined.
func NextItem(items []int64, states []ItemState, cursor int, forward bool) (int, bool) {
	if len(items) == 0 || len(states) != len(items) {
		return 0, true
	}

	// First pass: walk forward through pending items.
	if forward {
		for i := cursor + 1; i < len(items); i++ {
			if states[i] == StatePending {
				return i, false
			}
		}
		// Loop-back pass: cycle through skipped items.
		for i := range items {
			if states[i] == StateSkipped {
				return i, false
			}
		}
		return 0, true
	}

	// Backward (user pressed Previous): walk backward to find any non-completed item.
	for i := cursor - 1; i >= 0; i-- {
		if states[i] != StateCompleted {
			return i, false
		}
	}
	return 0, true
}
