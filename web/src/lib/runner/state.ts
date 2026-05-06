// Runner state machine: walks a routine's items, tracking per-item set progress.
//
// Two modes:
//   - sequential: complete all sets of one item, then move on
//   - alternating: round-robin one set at a time across items (circuit / superset)
//
// Skip semantics: skipping an item marks it 'skipped' but does NOT change
// setsDone — set progress is preserved. The next presentation prefers
// non-skipped items first; once everything else is exhausted, skipped items
// loop back so the user can finish them. Visiting a skipped item un-skips it.
//
// An item is "available" iff setsDone < totalSets (more sets remain). The
// routine is done when no item is available.

export type RunMode = 'sequential' | 'alternating';

export interface ItemRunState {
  setsDone: number;
  totalSets: number;
  skipped: boolean;
}

function available(s: ItemRunState): boolean {
  return s.setsDone < s.totalSets;
}
function preferred(s: ItemRunState): boolean {
  return available(s) && !s.skipped;
}

/** Compute the next item to present after a set was completed or skipped at `cursor`.
 *  Returns -1 when the routine is complete. */
export function nextItem(items: ItemRunState[], cursor: number, mode: RunMode): number {
  const n = items.length;
  if (n === 0) return -1;

  if (mode === 'sequential') {
    // Stay on cursor if it still has work and isn't currently flagged skipped.
    if (cursor >= 0 && cursor < n && preferred(items[cursor])) return cursor;

    // Forward sweep for preferred (non-skipped) items.
    for (let i = cursor + 1; i < n; i++) if (preferred(items[i])) return i;
    for (let i = 0; i <= cursor && i < n; i++) if (preferred(items[i])) return i;

    // Loop back: pick up skipped-but-still-available items in routine order.
    for (let i = cursor + 1; i < n; i++) if (available(items[i])) return i;
    for (let i = 0; i <= cursor && i < n; i++) if (available(items[i])) return i;

    return -1;
  }

  // Alternating: round-robin from cursor+1 — preferred first, then skipped.
  for (let offset = 1; offset <= n; offset++) {
    const i = (cursor + offset) % n;
    if (preferred(items[i])) return i;
  }
  for (let offset = 1; offset <= n; offset++) {
    const i = (cursor + offset) % n;
    if (available(items[i])) return i;
  }
  return -1;
}

/** Walk backward to find an item with sets remaining; returns -1 if none. */
export function previousItem(items: ItemRunState[], cursor: number): number {
  for (let i = cursor - 1; i >= 0; i--) {
    if (available(items[i])) return i;
  }
  return -1;
}
