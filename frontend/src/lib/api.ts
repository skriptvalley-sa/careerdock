// Centralised API client for the CareerDock backend.
// All requests use credentials: 'include' for httpOnly cookie auth.
// Includes automatic 401 → refresh → retry logic.

import type { DataResponse, ErrorBody, PaginatedResponse } from '@/types/api';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

/**
 * Custom error thrown on non-ok API responses.
 */
export class ApiError extends Error {
  status: number;
  code: string;
  details?: Record<string, unknown>;

  constructor(status: number, body: ErrorBody) {
    super(body.message);
    this.name = 'ApiError';
    this.status = status;
    this.code = body.code;
    this.details = body.details;
  }
}

// --- Token refresh lock ---
// Prevents concurrent refresh requests when multiple 401s arrive simultaneously.
let refreshPromise: Promise<boolean> | null = null;

/**
 * Attempt to refresh the access token. Returns true if successful.
 * Serialises concurrent calls so only one refresh request is in-flight at a time.
 */
async function refreshAccessToken(): Promise<boolean> {
  if (refreshPromise) {
    return refreshPromise;
  }

  refreshPromise = (async () => {
    try {
      const resp = await fetch(`${API_BASE}/api/auth/refresh`, {
        method: 'POST',
        credentials: 'include',
      });
      return resp.ok;
    } catch {
      return false;
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

// Paths that should never trigger auto-refresh (to avoid infinite loops).
const NO_REFRESH_PATHS = ['/api/auth/refresh', '/api/auth/login', '/api/auth/register'];

/**
 * Core fetch wrapper with 401 auto-refresh. On a 401 response, it attempts
 * one token refresh and retries the original request. If refresh fails, it
 * clears auth state via the onAuthFailure callback and throws the original error.
 */
async function fetchWithAuth(
  url: string,
  init: RequestInit,
  path: string,
): Promise<Response> {
  const resp = await fetch(url, { ...init, credentials: 'include' });

  // If not a 401, or if this is an auth endpoint, return as-is
  if (resp.status !== 401 || NO_REFRESH_PATHS.some((p) => path.startsWith(p))) {
    return resp;
  }

  // Attempt refresh
  const refreshed = await refreshAccessToken();
  if (!refreshed) {
    // Refresh failed — notify auth store to clear state
    if (onAuthFailure) onAuthFailure();
    return resp;
  }

  // Retry original request with fresh token
  return fetch(url, { ...init, credentials: 'include' });
}

// --- Auth failure callback ---
// Set by the AuthProvider to clear Zustand state when refresh fails.
let onAuthFailure: (() => void) | null = null;

/**
 * Register a callback invoked when token refresh fails (session expired).
 * Called from AuthProvider to wire up Zustand store clearing.
 */
export function setAuthFailureHandler(handler: () => void) {
  onAuthFailure = handler;
}

/**
 * Core fetch wrapper. Handles JSON encoding/decoding, error mapping,
 * and httpOnly cookie credentials with automatic 401 refresh/retry.
 */
async function api<T>(
  path: string,
  options: RequestInit & { params?: Record<string, string> } = {},
): Promise<T> {
  const { params, ...init } = options;

  let url = `${API_BASE}${path}`;
  if (params) {
    const searchParams = new URLSearchParams(params);
    url += `?${searchParams.toString()}`;
  }

  const headers: Record<string, string> = {
    ...(init.headers as Record<string, string>),
  };

  // Default to JSON content type for non-FormData bodies
  if (init.body && !(init.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  const resp = await fetchWithAuth(url, { ...init, headers }, path);

  // 204 No Content
  if (resp.status === 204) {
    return undefined as T;
  }

  const json = await resp.json();

  if (!resp.ok) {
    const errorBody = json.error as ErrorBody;
    throw new ApiError(resp.status, errorBody);
  }

  // Unwrap { data: T } envelope
  return (json as DataResponse<T>).data;
}

/**
 * Raw fetch wrapper — returns the full JSON body without unwrapping.
 * Used for paginated responses which have {data: [], pagination: {}}.
 */
async function apiRaw<T>(
  path: string,
  options: RequestInit & { params?: Record<string, string> } = {},
): Promise<T> {
  const { params, ...init } = options;

  let url = `${API_BASE}${path}`;
  if (params) {
    const searchParams = new URLSearchParams(params);
    url += `?${searchParams.toString()}`;
  }

  const headers: Record<string, string> = {
    ...(init.headers as Record<string, string>),
  };

  if (init.body && !(init.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  const resp = await fetchWithAuth(url, { ...init, headers }, path);

  const json = await resp.json();

  if (!resp.ok) {
    const errorBody = json.error as ErrorBody;
    throw new ApiError(resp.status, errorBody);
  }

  return json as T;
}

/**
 * Convenience methods matching REST verbs.
 */
export const apiClient = {
  get<T>(path: string, params?: Record<string, string>): Promise<T> {
    return api<T>(path, { method: 'GET', params });
  },

  post<T>(path: string, body?: unknown): Promise<T> {
    return api<T>(path, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    });
  },

  put<T>(path: string, body?: unknown): Promise<T> {
    return api<T>(path, {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    });
  },

  delete<T>(path: string, body?: unknown): Promise<T> {
    return api<T>(path, {
      method: 'DELETE',
      body: body ? JSON.stringify(body) : undefined,
    });
  },

  /** Fetch paginated response without unwrapping the data envelope. */
  getPaginated<T>(
    path: string,
    params?: Record<string, string>,
  ): Promise<PaginatedResponse<T>> {
    return apiRaw<PaginatedResponse<T>>(path, { method: 'GET', params });
  },

  /**
   * Fetch raw JSON without unwrapping — for endpoints that return non-standard
   * envelopes (e.g. admin list endpoints: {data: T[], total: N}).
   * Includes the same 401 auto-refresh logic as all other methods.
   */
  getRaw<T>(path: string, params?: Record<string, string>): Promise<T> {
    return apiRaw<T>(path, { method: 'GET', params });
  },

  upload<T>(path: string, formData: FormData): Promise<T> {
    return api<T>(path, {
      method: 'POST',
      body: formData,
      headers: {}, // Let browser set Content-Type with boundary
    });
  },
};
