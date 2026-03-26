'use client';

import { useState, useMemo, useCallback, useRef } from 'react';
import Link from 'next/link';
import { useParams, useRouter } from 'next/navigation';
import { Plus, ExternalLink } from 'lucide-react';
import {
  useListDetail,
  useUpdateEntry,
  useSyncListEntries,
} from '@/hooks/use-lists';
import { useUnsavedChangesGuard } from '@/hooks/use-unsaved-changes-guard';
import { CompanyStatusBadge, ALL_COMPANY_STATUSES, getCompanyStatusLabel } from '@/components/lists/company-status-badge';
import { CompanyBrowserPanel } from '@/components/lists/company-browser';
import { AddApplicationModal } from '@/components/lists/add-application-modal';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';
import type { CompanyTrackingStatus, ListEntry } from '@/types/api';

export default function ListDetailPage() {
  const params = useParams();
  const router = useRouter();
  const listId = params.id as string;
  const { data: list, isLoading } = useListDetail(listId);
  const updateEntry = useUpdateEntry();
  const syncEntries = useSyncListEntries();

  const [showBrowser, setShowBrowser] = useState(false);
  const [editingCompanyStatus, setEditingCompanyStatus] = useState<string | null>(null);
  const [applicationModalCompany, setApplicationModalCompany] = useState<{ id: string; name: string } | null>(null);

  // Track unsaved changes from CompanyBrowserPanel
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  // Use a ref for selected IDs — always up-to-date, no render-cycle lag
  const selectedIdsRef = useRef<string[]>([]);

  // Navigation guard — active when editing with unsaved changes
  const {
    showConfirmDialog,
    pendingHref,
    guardedNavigate,
    guardedCancel,
    confirmLeave,
    cancelLeave,
  } = useUnsavedChangesGuard(showBrowser && hasUnsavedChanges);

  // Set of company IDs already in this list (for edit mode pre-selection)
  const existingCompanyIds = useMemo(() => {
    if (!list) return new Set<string>();
    return new Set(list.entries.map((e) => e.company_id));
  }, [list]);

  // --- Callbacks for CompanyBrowserPanel ---

  const handleDiffChange = useCallback((hasChanges: boolean) => {
    setHasUnsavedChanges(hasChanges);
  }, []);

  const handleSelectedChange = useCallback((selectedIds: string[]) => {
    selectedIdsRef.current = selectedIds;
  }, []);

  const handleSyncSave = async (companyIds: string[]) => {
    try {
      await syncEntries.mutateAsync({ listId, company_ids: companyIds });
      setShowBrowser(false);
      setHasUnsavedChanges(false);
    } catch {
      // Save failed — stay in edit mode, let React Query surface the error
    }
  };

  const handleCloseBrowser = () => {
    setShowBrowser(false);
    setHasUnsavedChanges(false);
  };

  // Cancel button — if changes exist, show confirm dialog; otherwise close
  const handleCancel = () => {
    if (hasUnsavedChanges) {
      guardedCancel();
    } else {
      handleCloseBrowser();
    }
  };

  // Header Save button — uses ref for always-current selected IDs
  const handleHeaderSave = () => {
    handleSyncSave(selectedIdsRef.current);
  };

  // Confirm dialog: "Save & Leave" — save then navigate or close
  const handleSaveAndLeave = async () => {
    try {
      await syncEntries.mutateAsync({ listId, company_ids: selectedIdsRef.current });
      setShowBrowser(false);
      setHasUnsavedChanges(false);
      if (pendingHref) {
        confirmLeave();
      } else {
        // Cancel-triggered (no pendingHref) — just close the dialog
        cancelLeave();
      }
    } catch {
      // Save failed — stay on page
      cancelLeave();
    }
  };

  // Confirm dialog: "Discard" — discard changes and navigate or close
  const handleDiscard = () => {
    setShowBrowser(false);
    setHasUnsavedChanges(false);
    if (pendingHref) {
      confirmLeave();
    } else {
      // Cancel-triggered — just close everything
      cancelLeave();
    }
  };

  const handleCompanyStatusChange = async (entry: ListEntry, newStatus: CompanyTrackingStatus) => {
    await updateEntry.mutateAsync({
      listId,
      entryId: entry.id,
      company_status: newStatus,
    });
    setEditingCompanyStatus(null);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-primary)] border-t-transparent" />
      </div>
    );
  }

  if (!list) {
    return (
      <div className="py-12 text-center">
        <p className="text-[var(--color-text-muted)]">List not found.</p>
        <button
          onClick={() => router.push('/lists')}
          className="mt-4 text-sm text-[var(--color-primary)] hover:text-[var(--color-primary)]/80"
        >
          Back to lists
        </button>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <button
            onClick={() => guardedNavigate('/lists')}
            className="mb-2 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
          >
            &larr; Back to lists
          </button>
          <h1 className="text-2xl font-bold text-[var(--color-text)]">{list.name}</h1>
          {list.description && (
            <p className="mt-1 text-sm text-[var(--color-text-muted)]">{list.description}</p>
          )}
        </div>

        {/* Button repurposing: Edit List ↔ Cancel + Save */}
        <div className="flex items-center gap-3">
          {showBrowser ? (
            <>
              <button
                type="button"
                onClick={handleCancel}
                className="rounded-md border border-edge px-4 py-2 text-sm font-medium text-[var(--color-text)] hover:bg-overlay"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleHeaderSave}
                disabled={!hasUnsavedChanges || syncEntries.isPending}
                className="btn-primary rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
              >
                {syncEntries.isPending ? 'Saving...' : 'Save Changes'}
              </button>
            </>
          ) : (
            <button
              onClick={() => setShowBrowser(true)}
              className="btn-primary rounded-md px-4 py-2 text-sm font-medium"
            >
              Edit List
            </button>
          )}
        </div>
      </div>

      {/* Company browser panel (edit mode) */}
      {showBrowser && (
        <CompanyBrowserPanel
          existingCompanyIds={existingCompanyIds}
          onSave={handleSyncSave}
          onCancel={handleCancel}
          isSaving={syncEntries.isPending}
          onDiffChange={handleDiffChange}
          onSelectedChange={handleSelectedChange}
        />
      )}

      {/* Entry table — company-centric columns */}
      {list.entries.length === 0 && !showBrowser ? (
        <div className="mt-8 rounded-lg border border-dashed border-edge p-12 text-center">
          <p className="text-sm text-[var(--color-text-muted)]">
            No companies yet. Click &quot;Edit List&quot; to start adding companies.
          </p>
        </div>
      ) : list.entries.length > 0 ? (
        <div className="mt-6 overflow-x-auto rounded-lg border border-edge bg-card">
          <table className="min-w-full divide-y divide-edge">
            <thead className="bg-overlay">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">
                  Company
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">
                  Status
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">
                  Applications
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">
                  Add
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-edge">
              {list.entries.map((entry) => (
                <tr key={entry.id} className="hover:bg-surface">
                  {/* Company name linked to profile */}
                  <td className="whitespace-nowrap px-4 py-3">
                    {entry.company_slug ? (
                      <Link
                        href={`/companies/${entry.company_slug}`}
                        className="text-sm font-medium text-[var(--color-primary)] hover:text-[var(--color-primary)]/80 hover:underline"
                      >
                        {entry.company_name || entry.company_id.slice(0, 8) + '...'}
                      </Link>
                    ) : (
                      <div className="text-sm font-medium text-[var(--color-text)]">
                        {entry.company_name || entry.company_id.slice(0, 8) + '...'}
                      </div>
                    )}
                  </td>

                  {/* Company tracking status (inline editable, syncs across all lists) */}
                  <td className="whitespace-nowrap px-4 py-3">
                    {editingCompanyStatus === entry.id ? (
                      <select
                        value={entry.company_status}
                        onChange={(e) =>
                          handleCompanyStatusChange(entry, e.target.value as CompanyTrackingStatus)
                        }
                        onBlur={() => setEditingCompanyStatus(null)}
                        autoFocus
                        className="rounded-md border border-edge-input bg-input px-2 py-1 text-xs text-[var(--color-text)] focus:border-[var(--color-primary)]/50 focus:outline-none"
                      >
                        {ALL_COMPANY_STATUSES.map((s) => (
                          <option key={s} value={s}>
                            {getCompanyStatusLabel(s)}
                          </option>
                        ))}
                      </select>
                    ) : (
                      <button onClick={() => setEditingCompanyStatus(entry.id)}>
                        <CompanyStatusBadge status={entry.company_status} />
                      </button>
                    )}
                  </td>

                  {/* Application count chip — links to applications page filtered by company */}
                  <td className="whitespace-nowrap px-4 py-3 text-center">
                    {entry.application_count > 0 ? (
                      <Link
                        href={`/applications?company=${entry.company_id}`}
                        className="inline-flex items-center gap-1 rounded-full bg-[var(--color-primary)]/10 px-2.5 py-0.5 text-xs font-medium text-[var(--color-primary)] hover:bg-[var(--color-primary)]/20 transition-colors"
                      >
                        {entry.application_count}
                        <ExternalLink className="h-3 w-3" />
                      </Link>
                    ) : (
                      <span className="text-xs text-[var(--color-text-muted)]">—</span>
                    )}
                  </td>

                  {/* + Add Application button */}
                  <td className="whitespace-nowrap px-4 py-3 text-center">
                    <button
                      onClick={() =>
                        setApplicationModalCompany({
                          id: entry.company_id,
                          name: entry.company_name,
                        })
                      }
                      className="inline-flex items-center gap-1 rounded-md border border-edge px-2.5 py-1 text-xs font-medium text-[var(--color-text-muted)] hover:border-[var(--color-primary)]/30 hover:bg-surface hover:text-[var(--color-primary)] transition-colors"
                    >
                      <Plus className="h-3 w-3" />
                      Add
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}

      {/* Add Application Modal */}
      {applicationModalCompany && (
        <AddApplicationModal
          company={applicationModalCompany}
          onClose={() => setApplicationModalCompany(null)}
        />
      )}

      {/* Unsaved changes confirmation dialog */}
      <ConfirmDialog
        open={showConfirmDialog}
        title="Unsaved changes"
        message="You have unsaved changes to this list. Would you like to save them before leaving?"
        confirmLabel="Save & Leave"
        secondaryLabel="Discard"
        cancelLabel="Stay"
        onConfirm={handleSaveAndLeave}
        onSecondary={handleDiscard}
        onCancel={cancelLeave}
      />
    </div>
  );
}
