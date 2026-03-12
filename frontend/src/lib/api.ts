// Centralised API client for the CareerDock backend.
// All requests use credentials: 'include' for httpOnly cookie auth.

import type { DataResponse, ErrorBody } from '@/types/api';

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

/**
 * Core fetch wrapper. Handles JSON encoding/decoding, error mapping,
 * and httpOnly cookie credentials.
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

  const resp = await fetch(url, {
    ...init,
    headers,
    credentials: 'include',
  });

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

  delete<T>(path: string): Promise<T> {
    return api<T>(path, { method: 'DELETE' });
  },

  upload<T>(path: string, formData: FormData): Promise<T> {
    return api<T>(path, {
      method: 'POST',
      body: formData,
      headers: {}, // Let browser set Content-Type with boundary
    });
  },
};
