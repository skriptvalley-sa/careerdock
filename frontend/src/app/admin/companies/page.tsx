'use client';

import { useState, useEffect } from 'react';
import { Plus, Pencil, Search, Trash2, CheckCircle2, AlertCircle } from 'lucide-react';
import { useCompanyList } from '@/hooks/use-companies';
import { CompanyModal } from '@/components/admin/company-modal';
import { useAdminDeleteCompany } from '@/hooks/use-admin';
import type { CompanyDetail, CompanyListItem } from '@/types/api';
import { apiClient } from '@/lib/api';

export default function AdminCompaniesPage() {
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [modalOpen, setModalOpen] = useState(false);
  const [editCompany, setEditCompany] = useState<CompanyDetail | null>(null);
  const [toast, setToast] = useState<{ type: 'success' | 'error'; message: string } | null>(null);
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);

  const deleteCompany = useAdminDeleteCompany();

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search), 300);
    return () => clearTimeout(timer);
  }, [search]);

  // Auto-dismiss toast after 4 seconds
  useEffect(() => {
    if (!toast) return;
    const timer = setTimeout(() => setToast(null), 4000);
    return () => clearTimeout(timer);
  }, [toast]);

  const { data, isLoading } = useCompanyList(
    { q: debouncedSearch || undefined, limit: '50' },
    { staleTime: 0 }, // Admin always wants fresh data
  );

  const companies = data?.pages.flatMap((page) => page.data) ?? [];

  const handleEdit = async (c: CompanyListItem) => {
    try {
      const detail = await apiClient.get<CompanyDetail>(`/api/companies/${c.slug}`);
      setEditCompany(detail);
      setModalOpen(true);
    } catch {
      setEditCompany(c as unknown as CompanyDetail);
      setModalOpen(true);
    }
  };

  const handleCreate = () => {
    setEditCompany(null);
    setModalOpen(true);
  };

  const handleDelete = async (companyId: string, companyName: string) => {
    if (deleteConfirmId !== companyId) {
      // First click: ask for confirmation
      setDeleteConfirmId(companyId);
      return;
    }
    // Second click: confirmed — proceed
    setDeleteConfirmId(null);
    try {
      await deleteCompany.mutateAsync(companyId);
      setToast({ type: 'success', message: `"${companyName}" deleted from directory.` });
    } catch (err) {
      setToast({ type: 'error', message: err instanceof Error ? err.message : 'Failed to delete company.' });
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold text-[var(--color-text)]">Companies</h1>
        <button onClick={handleCreate} className="btn-primary flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium">
          <Plus className="h-4 w-4" />
          Create
        </button>
      </div>

      {/* Toast notification */}
      {toast && (
        <div
          className={`mt-4 flex items-center gap-2 rounded-md border px-4 py-3 text-sm ${
            toast.type === 'success'
              ? 'border-[#39ff14]/30 bg-[#39ff14]/10 text-[#39ff14]'
              : 'border-red-500/30 bg-red-500/10 text-red-400'
          }`}
        >
          {toast.type === 'success' ? (
            <CheckCircle2 className="h-4 w-4 shrink-0" />
          ) : (
            <AlertCircle className="h-4 w-4 shrink-0" />
          )}
          {toast.message}
        </div>
      )}

      <div className="relative mt-4">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--color-text-muted)]" />
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search companies..."
          className="block w-full rounded-md border border-edge-input bg-input py-2 pl-10 pr-3 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30"
        />
      </div>

      {/* Dismiss delete confirm on outside click */}
      {deleteConfirmId && (
        <div className="fixed inset-0 z-10" onClick={() => setDeleteConfirmId(null)} />
      )}

      <div className="mt-4 overflow-x-auto rounded-lg border border-edge">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-edge bg-overlay text-left">
              <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Name</th>
              <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Slug</th>
              <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Size</th>
              <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">HQ</th>
              <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Hiring</th>
              <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Tech Stack</th>
              <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]" />
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-[var(--color-text-muted)]">
                  <div className="flex items-center justify-center">
                    <div className="h-5 w-5 animate-spin rounded-full border-2 border-[var(--color-primary)] border-t-transparent" />
                  </div>
                </td>
              </tr>
            ) : companies.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-[var(--color-text-muted)]">
                  No companies found.
                </td>
              </tr>
            ) : (
              companies.map((c) => (
                <tr
                  key={c.id}
                  className="border-b border-edge hover:bg-card/50"
                  onClick={() => deleteConfirmId && setDeleteConfirmId(null)}
                >
                  <td className="px-4 py-3 font-medium text-[var(--color-text)]">{c.name}</td>
                  <td className="px-4 py-3 text-[var(--color-text-muted)]">{c.slug}</td>
                  <td className="px-4 py-3 text-[var(--color-text-muted)]">{c.size || '-'}</td>
                  <td className="px-4 py-3 text-[var(--color-text-muted)]">{c.headquarters || '-'}</td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                        c.hiring_status === 'active'
                          ? 'bg-green-500/15 text-green-400'
                          : c.hiring_status === 'paused'
                            ? 'bg-amber-500/15 text-amber-400'
                            : 'bg-slate-500/15 text-[var(--color-text-muted)]'
                      }`}
                    >
                      {c.hiring_status}
                    </span>
                  </td>
                  <td className="max-w-[200px] truncate px-4 py-3 text-[var(--color-text-muted)]">
                    {c.tech_stack.slice(0, 3).join(', ')}
                    {c.tech_stack.length > 3 && ` +${c.tech_stack.length - 3}`}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
                      <button
                        onClick={() => handleEdit(c)}
                        className="rounded-md p-1.5 text-[var(--color-text-muted)] hover:bg-card hover:text-[var(--color-text)]"
                        title="Edit"
                      >
                        <Pencil className="h-4 w-4" />
                      </button>
                      {deleteConfirmId === c.id ? (
                        <button
                          onClick={() => handleDelete(c.id, c.name)}
                          disabled={deleteCompany.isPending}
                          className="z-20 rounded-md bg-red-500/20 px-2 py-1 text-xs font-medium text-red-400 hover:bg-red-500/30 disabled:opacity-50"
                          title="Click again to confirm deletion"
                        >
                          Confirm delete?
                        </button>
                      ) : (
                        <button
                          onClick={() => handleDelete(c.id, c.name)}
                          className="rounded-md p-1.5 text-[var(--color-text-muted)] hover:bg-card hover:text-red-400"
                          title="Delete company"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {modalOpen && (
        <CompanyModal
          company={editCompany}
          onClose={() => {
            setModalOpen(false);
            setEditCompany(null);
          }}
          onSuccess={(msg) => setToast({ type: 'success', message: msg })}
        />
      )}
    </div>
  );
}
