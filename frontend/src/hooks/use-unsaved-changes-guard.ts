'use client';

import { useState, useCallback, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';

interface UseUnsavedChangesGuardReturn {
  /** Whether the confirm dialog should be shown */
  showConfirmDialog: boolean;
  /** The href the user tried to navigate to (null if cancel-triggered) */
  pendingHref: string | null;
  /** Call this instead of router.push() — shows dialog if changes exist */
  guardedNavigate: (href: string) => void;
  /** Call when a cancel/close action should be guarded too */
  guardedCancel: () => void;
  /** User confirmed they want to leave — navigates to pendingHref */
  confirmLeave: () => void;
  /** User chose to stay — closes the dialog */
  cancelLeave: () => void;
}

/**
 * Hook that guards against losing unsaved changes.
 * Handles:
 * 1. Browser refresh / tab close (beforeunload — native browser dialog)
 * 2. Browser back/forward (popstate interception)
 * 3. Explicit guardedNavigate / guardedCancel calls
 */
export function useUnsavedChangesGuard(enabled: boolean): UseUnsavedChangesGuardReturn {
  const router = useRouter();
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const [pendingHref, setPendingHref] = useState<string | null>(null);

  // Ref to avoid stale closures in confirmLeave
  const pendingHrefRef = useRef<string | null>(null);

  // Keep ref in sync with state
  useEffect(() => {
    pendingHrefRef.current = pendingHref;
  }, [pendingHref]);

  // --- Browser beforeunload (refresh, close tab) ---
  useEffect(() => {
    if (!enabled) return;
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault();
      e.returnValue = '';
    };
    window.addEventListener('beforeunload', handler);
    return () => window.removeEventListener('beforeunload', handler);
  }, [enabled]);

  // --- Browser back/forward (popstate) ---
  useEffect(() => {
    if (!enabled) return;

    const handler = () => {
      // Re-push current URL to prevent actual navigation
      window.history.pushState(null, '', window.location.href);

      setPendingHref(null);
      setShowConfirmDialog(true);
    };

    window.addEventListener('popstate', handler);
    return () => window.removeEventListener('popstate', handler);
  }, [enabled]);

  // Guarded in-app navigation
  const guardedNavigate = useCallback(
    (href: string) => {
      if (enabled) {
        setPendingHref(href);
        setShowConfirmDialog(true);
      } else {
        router.push(href);
      }
    },
    [enabled, router],
  );

  // Guarded cancel/close action
  const guardedCancel = useCallback(() => {
    if (enabled) {
      setPendingHref(null);
      setShowConfirmDialog(true);
    }
  }, [enabled]);

  // User confirmed leaving
  const confirmLeave = useCallback(() => {
    setShowConfirmDialog(false);
    const href = pendingHrefRef.current;
    setPendingHref(null);
    if (href) {
      router.push(href);
    }
  }, [router]);

  // User chose to stay
  const cancelLeave = useCallback(() => {
    setShowConfirmDialog(false);
    setPendingHref(null);
  }, []);

  return {
    showConfirmDialog,
    pendingHref,
    guardedNavigate,
    guardedCancel,
    confirmLeave,
    cancelLeave,
  };
}
