'use client';

import { useState, useMemo } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { ChevronDown } from 'lucide-react';
import { useAllEntries, useUpdateEntry } from '@/hooks/use-lists';
import { StatusBadge, ALL_STATUSES, getStatusLabel } from '@/components/lists/status-badge';
import type { ApplicationStatus, ListEntry } from '@/types/api';

export default function ApplicationsPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const initialStatus = searchParams.get('status') || '';
  const [statusFilter, setStatusFilter] = useState(initialStatus);
  const [companyFilter, setCompanyFilter] = useState('');
  const [editingEntry, setEditingEntry] = useState<string | null>(null);

  const { data: entries, isLoading } = useAllEntries(statusFilter || undefined);
  const updateEntry = useUpdateEntry();

  // Extract unique companies from entries for the company filter dropdown
  const uniqueCompanies = useMemo(() => {
    if (!entries) return [];
    const map = new Map<string, string>();
    for (const e of entries) {
      if (e.company_id && e.company_name && !map.has(e.company_id)) {
        map.set(e.company_id, e.company_name);
      }
    }
    return Array.from(map.entries())
      .map(([id, name]) => ({ id, name }))
      .sort((a, b) => a.name.localeCompare(b.name));
  }, [entries]);

  // Apply client-side company filter
  const filteredEntries = useMemo(() => {
    if (!entries) return undefined;
    if (!companyFilter) return entries;
    return entries.filter((e) => e.company_id === companyFilter);
  }, [entries, companyFilter]);

  const handleStatusChange = async (
    entry: ListEntry & { list_name: string; company_name: string },
    newStatus: ApplicationStatus,
  ) => {
    await updateEntry.mutateAsync({
      listId: entry.list_id,
      entryId: entry.id,
      status: newStatus,
    });
    setEditingEntry(null);
  };

  const handleFilterChange = (status: string) => {
    setStatusFilter(status);
    if (status) {
      router.replace(`/applications?status=${status}`, { scroll: false });
    } else {
      router.replace('/applications', { scroll: false });
    }
  };

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-slate-100">All Applications</h1>
        <p className="mt-1 text-sm text-slate-500">
          Track all your applications across every list in one place.
        </p>
      </div>

      {/* Filters row: status chips + company dropdown */}
      <div className="mb-6 flex flex-wrap items-center gap-3">
        {/* Status filter chips */}
        <div className="flex flex-wrap items-center gap-2">
          <button
            type="button"
            onClick={() => handleFilterChange('')}
            className={`rounded-full border px-3 py-1 text-xs font-medium transition-all ${
              !statusFilter
                ? 'border-[#00f0ff]/50 bg-[#00f0ff]/10 text-[#00f0ff]'
                : 'border-edge text-slate-600 hover:border-edge-hover hover:text-slate-400'
            }`}
          >
            All
          </button>
          {ALL_STATUSES.filter((s) => s !== 'not_applied').map((s) => (
            <button
              key={s}
              type="button"
              onClick={() => handleFilterChange(s)}
              className={`rounded-full border px-3 py-1 text-xs font-medium transition-all ${
                statusFilter === s
                  ? 'border-[#00f0ff]/50 bg-[#00f0ff]/10 text-[#00f0ff]'
                  : 'border-edge text-slate-600 hover:border-edge-hover hover:text-slate-400'
              }`}
            >
              {getStatusLabel(s)}
            </button>
          ))}
        </div>

        {/* Company filter dropdown */}
        {uniqueCompanies.length > 0 && (
          <div className="relative">
            <select
              value={companyFilter}
              onChange={(e) => setCompanyFilter(e.target.value)}
              className="appearance-none rounded-full border border-edge bg-overlay py-1 pl-3 pr-8 text-xs font-medium text-slate-400 transition-all hover:border-edge-hover hover:text-slate-300 focus:border-[#00f0ff]/50 focus:text-[#00f0ff] focus:outline-none"
            >
              <option value="">All Companies</option>
              {uniqueCompanies.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
            <ChevronDown className="pointer-events-none absolute right-2.5 top-1/2 h-3 w-3 -translate-y-1/2 text-slate-500" />
          </div>
        )}
      </div>

      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-[#00f0ff] border-t-transparent" />
        </div>
      )}

      {!isLoading && filteredEntries && filteredEntries.length === 0 && (
        <div className="rounded-lg border border-dashed border-edge p-12 text-center">
          <p className="text-sm text-slate-500">
            {statusFilter || companyFilter
              ? 'No applications match the selected filters.'
              : 'No applications yet. Add entries to your lists to see them here.'}
          </p>
        </div>
      )}

      {filteredEntries && filteredEntries.length > 0 && (
        <div className="overflow-hidden rounded-lg border border-edge bg-card">
          <table className="min-w-full divide-y divide-edge">
            <thead className="bg-overlay">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                  Company
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                  Role
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                  Status
                </th>
                <th className="hidden px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500 sm:table-cell">
                  List
                </th>
                <th className="hidden px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500 md:table-cell">
                  Date Applied
                </th>
                <th className="hidden px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500 lg:table-cell">
                  Notes
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-edge">
              {filteredEntries.map((entry) => (
                <tr key={entry.id} className="hover:bg-surface">
                  <td className="whitespace-nowrap px-4 py-3">
                    <div className="text-sm font-medium text-slate-100">
                      {entry.company_name || entry.company_id.slice(0, 8) + '...'}
                    </div>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3">
                    <div className="text-sm text-slate-300">{entry.role_title || '-'}</div>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3">
                    {editingEntry === entry.id ? (
                      <select
                        value={entry.status}
                        onChange={(e) =>
                          handleStatusChange(entry, e.target.value as ApplicationStatus)
                        }
                        onBlur={() => setEditingEntry(null)}
                        autoFocus
                        className="rounded-md border border-edge-input bg-input px-2 py-1 text-xs text-slate-200 focus:border-[#00f0ff]/50 focus:outline-none"
                      >
                        {ALL_STATUSES.filter((s) => s !== 'not_applied').map((s) => (
                          <option key={s} value={s}>
                            {getStatusLabel(s)}
                          </option>
                        ))}
                      </select>
                    ) : (
                      <button onClick={() => setEditingEntry(entry.id)}>
                        <StatusBadge status={entry.status} />
                      </button>
                    )}
                  </td>
                  <td className="hidden whitespace-nowrap px-4 py-3 text-sm text-slate-500 sm:table-cell">
                    {entry.list_name}
                  </td>
                  <td className="hidden whitespace-nowrap px-4 py-3 text-sm text-slate-500 md:table-cell">
                    {entry.date_applied || '-'}
                  </td>
                  <td className="hidden max-w-xs truncate px-4 py-3 text-sm text-slate-500 lg:table-cell">
                    {entry.notes || '-'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {filteredEntries && filteredEntries.length > 0 && (
        <div className="mt-4 text-xs text-slate-600">
          {filteredEntries.length} application{filteredEntries.length !== 1 ? 's' : ''}
          {statusFilter ? ` with status "${getStatusLabel(statusFilter as ApplicationStatus)}"` : ''}
          {companyFilter
            ? ` at ${uniqueCompanies.find((c) => c.id === companyFilter)?.name ?? 'selected company'}`
            : ''}
        </div>
      )}
    </div>
  );
}
