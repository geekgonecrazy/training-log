// Tiny typed REST client for the gRPC-Gateway HTTP/JSON surface.
//
// All requests use credentials: 'include' so the access_token cookie travels
// automatically. On 401 we transparently call /v1/auth/refresh once and retry.
// A second 401 throws an UnauthenticatedError which the auth store turns into
// a redirect to /login.

export class ApiError extends Error {
  constructor(public status: number, public code: string, message: string) {
    super(message);
  }
}

export class UnauthenticatedError extends ApiError {
  constructor(message = 'unauthenticated') {
    super(401, 'UNAUTHENTICATED', message);
  }
}

let refreshing: Promise<boolean> | null = null;

async function refreshOnce(): Promise<boolean> {
  if (!refreshing) {
    refreshing = (async () => {
      const res = await fetch('/v1/auth/refresh', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: '{}'
      });
      return res.ok;
    })().finally(() => {
      // Allow a fresh refresh attempt next time.
      setTimeout(() => (refreshing = null), 0);
    });
  }
  return refreshing;
}

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  body?: unknown;
  // If true, never attempt a refresh on 401. Used by /v1/auth/* itself.
  skipAuthRefresh?: boolean;
}

export async function api<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const init: RequestInit = {
    method: opts.method ?? 'GET',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' }
  };
  if (opts.body !== undefined) {
    init.body = JSON.stringify(opts.body);
  }

  let res = await fetch(path, init);

  if (res.status === 401 && !opts.skipAuthRefresh && !path.startsWith('/v1/auth/')) {
    const ok = await refreshOnce();
    if (ok) {
      res = await fetch(path, init);
    }
  }

  if (res.status === 401) {
    throw new UnauthenticatedError();
  }

  if (!res.ok) {
    let body: { code?: string; message?: string } = {};
    try {
      body = await res.json();
    } catch {
      // ignore body parse errors
    }
    throw new ApiError(res.status, body.code ?? 'UNKNOWN', body.message ?? res.statusText);
  }

  if (res.status === 204) {
    return undefined as T;
  }
  return (await res.json()) as T;
}
