'use client';

import { useEffect, useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/store/auth-store';
import { queryKeys } from '@/lib/query-keys';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

/**
 * useSSE connects to the backend Server-Sent Events endpoint and
 * automatically invalidates TanStack Query caches when events arrive.
 *
 * Currently handled events:
 * - resume_ready: invalidates resumes list + credits
 */
export function useSSE() {
  const { isAuthenticated } = useAuthStore();
  const qc = useQueryClient();
  const esRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!isAuthenticated) {
      // Close any existing connection when logged out
      if (esRef.current) {
        esRef.current.close();
        esRef.current = null;
      }
      return;
    }

    // Don't create duplicate connections
    if (esRef.current) return;

    const es = new EventSource(`${API_BASE}/api/events`, {
      withCredentials: true,
    });
    esRef.current = es;

    es.addEventListener('connected', () => {
      // Connection established — no action needed
    });

    es.addEventListener('resume_ready', () => {
      // Resume processing complete — refresh resume list and credits
      qc.invalidateQueries({ queryKey: queryKeys.resumes.all });
      qc.invalidateQueries({ queryKey: queryKeys.credits.balance });
    });

    es.onerror = () => {
      // EventSource auto-reconnects on error; just close stale ref
      // so next mount recreates it cleanly
      es.close();
      esRef.current = null;
    };

    return () => {
      es.close();
      esRef.current = null;
    };
  }, [isAuthenticated, qc]);
}
