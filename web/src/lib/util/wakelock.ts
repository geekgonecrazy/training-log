// Tiny wrapper around the Screen Wake Lock API. Browsers auto-release the
// lock when the page is hidden, so we re-acquire on visibilitychange while
// the caller still wants the screen on.
//
// Usage:
//   const lock = acquireWakeLock();   // request + auto re-acquire on focus
//   ...
//   await lock.release();             // when done
//
// Safe to call on browsers that don't support Wake Lock — request() is a no-op
// and release() resolves immediately.

interface WakeLockHandle {
  release(): Promise<void>;
}

interface WakeLockSentinelLike {
  released?: boolean;
  release(): Promise<void>;
  addEventListener(type: 'release', listener: () => void): void;
}

interface WakeLockApi {
  request(type: 'screen'): Promise<WakeLockSentinelLike>;
}

export function acquireWakeLock(): WakeLockHandle {
  const wakeLock: WakeLockApi | undefined =
    typeof navigator !== 'undefined'
      ? (navigator as Navigator & { wakeLock?: WakeLockApi }).wakeLock
      : undefined;
  if (!wakeLock) {
    return { release: async () => {} };
  }

  let sentinel: WakeLockSentinelLike | null = null;
  let cancelled = false;

  async function request() {
    if (cancelled) return;
    try {
      sentinel = await wakeLock!.request('screen');
      sentinel.addEventListener('release', () => { sentinel = null; });
    } catch {
      // Permission denied / page not visible — best effort.
      sentinel = null;
    }
  }

  function onVisibility() {
    if (!cancelled && document.visibilityState === 'visible' && !sentinel) {
      void request();
    }
  }

  void request();
  document.addEventListener('visibilitychange', onVisibility);

  return {
    async release() {
      cancelled = true;
      document.removeEventListener('visibilitychange', onVisibility);
      try { await sentinel?.release(); } catch { /* ignore */ }
      sentinel = null;
    }
  };
}
