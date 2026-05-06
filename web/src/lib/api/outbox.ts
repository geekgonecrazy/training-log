// Offline write queue for session logs.
//
// Why only sessions: machines/exercises/routines are configuration changes
// that should fail loud if offline; sessions are the high-volume in-the-moment
// writes that can't be lost. Sessions carry a clientId UUID, so the server
// safely de-dupes on retry.

import Dexie, { type Table } from 'dexie';
import { sessions, type SessionInput } from './endpoints';
import { ApiError } from './client';

interface OutboxRow {
  id?: number;
  payload: SessionInput;
  createdAt: number;
}

class OutboxDB extends Dexie {
  outbox!: Table<OutboxRow, number>;
  constructor() {
    super('training-log-outbox');
    this.version(1).stores({ outbox: '++id, createdAt' });
  }
}

const db = new OutboxDB();

export async function logSessionResilient(payload: SessionInput): Promise<void> {
  try {
    await sessions.log(payload);
    return;
  } catch (err) {
    if (err instanceof ApiError && err.status >= 400 && err.status < 500 && err.status !== 401) {
      // 4xx (except auth) means the request is malformed — don't queue garbage.
      throw err;
    }
    // Network failure, server down, or auth issue → queue for later.
    await db.outbox.add({ payload, createdAt: Date.now() });
  }
}

export async function flushOutbox(): Promise<{ flushed: number; remaining: number }> {
  const rows = await db.outbox.orderBy('createdAt').toArray();
  let flushed = 0;
  for (const row of rows) {
    try {
      await sessions.log(row.payload);
      if (row.id !== undefined) {
        await db.outbox.delete(row.id);
        flushed++;
      }
    } catch (err) {
      if (err instanceof ApiError && err.status >= 400 && err.status < 500 && err.status !== 401) {
        // Permanent failure — drop it.
        if (row.id !== undefined) await db.outbox.delete(row.id);
      } else {
        // Transient — stop and retry later in arrival order.
        break;
      }
    }
  }
  const remaining = await db.outbox.count();
  return { flushed, remaining };
}

export async function outboxCount(): Promise<number> {
  return db.outbox.count();
}

let setupDone = false;
export function setupOutboxAutoFlush() {
  if (setupDone || typeof window === 'undefined') return;
  setupDone = true;
  window.addEventListener('online', () => {
    void flushOutbox();
  });
  // Also flush on tab focus.
  window.addEventListener('focus', () => {
    if (navigator.onLine) void flushOutbox();
  });
}
