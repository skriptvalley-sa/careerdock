'use client';

import { useState, useMemo } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { ChevronDown } from 'lucide-react';
import { useApplications, useUpdateApplication } from '@/hooks/use-applications';
import { StatusBadge, ALL_STATUSES, getStatusLabel } from '@/components/lists/status-badge';
import type { Application, ApplicationStatus } from '@/types/api';

export default function ApplicationsPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const initialStatus = searchParams.get('status') || '';
  const [statusFilter, setStatusFilter] = useState(initialStatus);
  const [companyFilter, setCompanyFilter] = useState('');
  const [editingEntry, setEditingEntry] = useState<string | null>(null);

  const { data: applications, isLoading } = useApplications(statusFilter || undefined);
  const updateApp = useUpdateApplication();

  // Extract unique companies for the company filter dropdown
  const uniqueCompanies = useMemo(() => {
    if (!applications) return [];
    const map = new Map<string, string>();
    for (const a of applications) {
      if (a.company_id && a.company_name && !map.has(a.company_id)) {
        map.set(a.company_id, a.company_name);
      }
    }
    return Array.from(map.entries())
      .map(([id, name]) => ({ id, name }))
      .sort((a, b) => a.name.localeCompare(b.name));
  }, [applications]);

  // Apply client-side company filter
  const filtered = useMemo(() => {
    if (!applications) return undefined;
    if (!companyFilter) return applications;
    return applications.filter((a) => a.company_id === companyFilter);
  }, [applications, companyFilter]);

  const handleStatusChange = async (app: Application, newStatus: ApplicationStatus) => {
    await updateApp.mutateAsync({ id: app.id, status: newStatus });
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
          Track all your applications across every company in one place.
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

      {!isLoading && filtered && filtered.length === 0 && (
        <div className="rounded-lg border border-dashed border-edge p-12 text-center">
          <p className="text-sm text-slate-500">
            {statusFilter || companyFilter
              ? 'No applications match the selected filters.'
              : 'No applications yet. Add applications from a list detail page.'}
          </p>
        </div>
      )}

      {filtered && filtered.length > 0 && (
        <div className="overflow-x-auto rounded-lg border border-edge bg-card">
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
                <th className="hidden px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500 md:table-cell">
                  Date Applied
                </th>
                <th className="hidden px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500 lg:table-cell">
                  Notes
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-edge">
              {filtered.map((app) => (
                <tr key={app.id} className="hover:bg-surface">
                  <td className="whitespace-nowrap px-4 py-3">
                    <div className="text-sm font-medium text-slate-100">
                      {app.company_name || app.company_id.slice(0, 8) + '...'}
                    </div>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3">
                    <div className="text-sm text-slate-300">{app.role_title || '-'}</div>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3">
                    {editingEntry === app.id ? (
                      <select
                        value={app.status}
                        onChange={(e) =>
                          handleStatusChange(app, e.target.value as ApplicationStatus)
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
                      <button onClick={() => setEditingEntry(app.id)}>
                        <StatusBadge status={app.status} />
                      </button>
                    )}
                  </td>
                  <td className="hidden whitespace-nowrap px-4 py-3 text-sm text-slate-500 md:table-cell">
                    {app.date_applied || '-'}
                  </td>
                  <td className="hidden max-w-xs truncate px-4 py-3 text-sm text-slate-500 lg:table-cell">
                    {app.notes || '-'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {filtered && filtered.length > 0 && (
        <div className="mt-4 text-xs text-slate-600">
          {filtered.length} application{filtered.length !== 1 ? 's' : ''}
          {statusFilter ? ` with status "${getStatusLabel(statusFilter as ApplicationStatus)}"` : ''}
          {companyFilter
            ? ` at ${uniqueCompanies.find((c) => c.id === companyFilter)?.name ?? 'selected company'}`
            : ''}
        </div>
      )}
    </div>
  );
}
