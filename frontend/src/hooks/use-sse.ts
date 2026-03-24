'use client';

import { useEffect, useRef, useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/store/auth-store';
import { queryKeys } from '@/lib/query-keys';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const MAX_RETRY_DELAY = 30_000; // 30 seconds
const INITIAL_RETRY_DELAY = 1_000; // 1 second

/**
 * useSSE connects to the backend Server-Sent Events endpoint and
 * automatically invalidates TanStack Query caches when events arrive.
 *
 * Reconnects automatically with exponential backoff on connection errors.
 *
 * Currently handled events:
 * - resume_ready: invalidates resumes list + credits
 */
export function useSSE() {
  const { isAuthenticated } = useAuthStore();
  const qc = useQueryClient();
  const esRef = useRef<EventSource | null>(null);
  const retryTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const retryDelayRef = useRef(INITIAL_RETRY_DELAY);

  const connect = useCallback(() => {
    // Clean up any existing connection
    if (esRef.current) {
      esRef.current.close();
      esRef.current = null;
    }

    const es = new EventSource(`${API_BASE}/api/events`, {
      withCredentials: true,
    });
    esRef.current = es;

    es.addEventListener('connected', () => {
      // Connection established — reset retry delay
      retryDelayRef.current = INITIAL_RETRY_DELAY;
    });

    es.addEventListener('resume_ready', () => {
      // Resume processing complete — refresh resume list and credits
      qc.invalidateQueries({ queryKey: queryKeys.resumes.all });
      qc.invalidateQueries({ queryKey: queryKeys.credits.balance });
    });

    es.onerror = () => {
      // Close the failed connection
      es.close();
      esRef.current = null;

      // Reconnect with exponential backoff
      const delay = retryDelayRef.current;
      retryDelayRef.current = Math.min(delay * 2, MAX_RETRY_DELAY);

      retryTimeoutRef.current = setTimeout(() => {
        retryTimeoutRef.current = null;
        // Only reconnect if still authenticated
        if (useAuthStore.getState().isAuthenticated) {
          connect();
        }
      }, delay);
    };
  }, [qc]);

  useEffect(() => {
    if (!isAuthenticated) {
      // Close any existing connection when logged out
      if (esRef.current) {
        esRef.current.close();
        esRef.current = null;
      }
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current);
        retryTimeoutRef.current = null;
      }
      retryDelayRef.current = INITIAL_RETRY_DELAY;
      return;
    }

    // Don't create duplicate connections
    if (esRef.current) return;

    connect();

    return () => {
      if (esRef.current) {
        esRef.current.close();
        esRef.current = null;
      }
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current);
        retryTimeoutRef.current = null;
      }
      retryDelayRef.current = INITIAL_RETRY_DELAY;
    };
  }, [isAuthenticated, connect]);
}
