'use client';

import { useState, useMemo } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft, ExternalLink, MapPin, Building2, Calendar, DollarSign, Briefcase, List } from 'lucide-react';
import { useCompanyDetail } from '@/hooks/use-companies';
import { useEntriesByCompany, useUpdateEntry } from '@/hooks/use-lists';
import { useAuthStore } from '@/store/auth-store';
import { TechStackTags } from '@/components/companies/tech-stack-tags';
import { StatusBadge, ALL_STATUSES, getStatusLabel } from '@/components/lists/status-badge';
import { CompanyStatusBadge } from '@/components/lists/company-status-badge';
import type { ApplicationStatus, CompanyTrackingStatus, ListEntry } from '@/types/api';

const sizeLabels: Record<string, string> = {
  startup: 'Startup',
  small: 'Small',
  mid: 'Mid-size',
  large: 'Large',
  enterprise: 'Enterprise',
};

const tierLabels: Record<string, string> = {
  tier_1: 'Tier 1 (>40L)',
  tier_2: 'Tier 2 (20-40L)',
  tier_3: 'Tier 3 (10-20L)',
  tier_4: 'Tier 4 (5-10L)',
};

const hiringColors: Record<string, string> = {
  active: 'bg-green-900/30 text-green-400 border-green-800',
  paused: 'bg-yellow-900/30 text-yellow-400 border-yellow-800',
  unknown: 'bg-slate-800 text-slate-500 border-edge',
};

// Company tracking status priority for "best" status display
const statusPriority: Record<CompanyTrackingStatus, number> = {
  marked: 0,
  researching: 1,
  applied: 2,
  interviewing: 3,
  offered: 4,
  accepted: 5,
  rejected: -1,
};

export default function CompanyProfile() {
  const { slug } = useParams<{ slug: string }>();
  const { data: company, isLoading, isError, error } = useCompanyDetail(slug);
  const { isAuthenticated } = useAuthStore();
  const { data: userEntries } = useEntriesByCompany(
    isAuthenticated && company ? company.id : undefined,
  );
  const updateEntry = useUpdateEntry();
  const [editingStatus, setEditingStatus] = useState<string | null>(null);

  // Derive overall company status (highest priority across all list entries)
  const overallStatus = useMemo<CompanyTrackingStatus | null>(() => {
    if (!userEntries || userEntries.length === 0) return null;
    let best: CompanyTrackingStatus = 'marked';
    let bestPriority = -2;
    for (const e of userEntries) {
      const p = statusPriority[e.company_status] ?? 0;
      if (p > bestPriority) {
        bestPriority = p;
        best = e.company_status;
      }
    }
    return best;
  }, [userEntries]);

  // Derive unique lists (for "Your Lists" chips)
  const uniqueLists = useMemo(() => {
    if (!userEntries) return [];
    const seen = new Map<string, { listId: string; listName: string }>();
    for (const e of userEntries) {
      if (!seen.has(e.list_id)) {
        seen.set(e.list_id, { listId: e.list_id, listName: e.list_name });
      }
    }
    return Array.from(seen.values());
  }, [userEntries]);

  // Filter entries to only those with actual applications (not just list additions)
  const appliedEntries = useMemo(() => {
    if (!userEntries) return [];
    return userEntries.filter((e) => e.status !== 'not_applied');
  }, [userEntries]);

  const handleStatusChange = async (
    entry: ListEntry & { list_name: string },
    newStatus: ApplicationStatus,
  ) => {
    await updateEntry.mutateAsync({
      listId: entry.list_id,
      entryId: entry.id,
      status: newStatus,
    });
    setEditingStatus(null);
  };

  if (isLoading) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-12 sm:px-6 lg:px-8">
        <div className="animate-pulse space-y-6">
          <div className="h-8 w-64 rounded bg-slate-800" />
          <div className="h-4 w-96 rounded bg-slate-800" />
          <div className="h-48 rounded-lg bg-slate-800" />
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-12 sm:px-6 lg:px-8">
        <div className="rounded-lg bg-red-900/30 p-6 text-center">
          <p className="text-sm text-red-400">
            Failed to load company: {(error as Error).message}
          </p>
          <Link href="/companies" className="mt-4 inline-block text-sm text-[#00f0ff] hover:text-[#00f0ff]/80">
            Back to directory
          </Link>
        </div>
      </div>
    );
  }

  if (!company) return null;

  return (
    <div className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
      {/* Breadcrumb */}
      <Link
        href="/companies"
        className="mb-6 inline-flex items-center gap-1.5 text-sm text-slate-500 hover:text-slate-300"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to directory
      </Link>

      {/* Header */}
      <div className="mb-8">
        <div className="flex items-start justify-between gap-4">
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-3xl font-bold text-slate-100">{company.name}</h1>
              {overallStatus && <CompanyStatusBadge status={overallStatus} />}
            </div>
            <div className="mt-2 flex flex-wrap items-center gap-3 text-sm text-slate-500">
              {company.headquarters && (
                <span className="flex items-center gap-1">
                  <MapPin className="h-4 w-4" />
                  {company.headquarters}
                </span>
              )}
              {company.size && (
                <span className="flex items-center gap-1">
                  <Building2 className="h-4 w-4" />
                  {sizeLabels[company.size] || company.size}
                </span>
              )}
              {company.founded_year && (
                <span className="flex items-center gap-1">
                  <Calendar className="h-4 w-4" />
                  Founded {company.founded_year}
                </span>
              )}
            </div>
          </div>
          <span
            className={`shrink-0 rounded-full border px-3 py-1 text-sm font-medium ${hiringColors[company.hiring_status] || hiringColors.unknown}`}
          >
            {company.hiring_status === 'active'
              ? 'Actively Hiring'
              : company.hiring_status === 'paused'
                ? 'Hiring Paused'
                : 'Status Unknown'}
          </span>
        </div>
      </div>

      {/* Description */}
      {company.description && (
        <section className="mb-8">
          <p className="text-slate-300 leading-relaxed">{company.description}</p>
        </section>
      )}

      {/* Key Info Grid */}
      <div className="mb-8 grid gap-6 sm:grid-cols-2">
        {/* Compensation */}
        <div className="rounded-lg border border-edge p-5">
          <h2 className="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-100">
            <DollarSign className="h-4 w-4" />
            Compensation
          </h2>
          <div className="space-y-2 text-sm">
            {company.compensation_tier && (
              <div className="flex justify-between">
                <span className="text-slate-500">Tier</span>
                <span className="font-medium text-slate-100">
                  {tierLabels[company.compensation_tier] || company.compensation_tier}
                </span>
              </div>
            )}
            <div className="flex justify-between">
              <span className="text-slate-500">RSU</span>
              <span className="font-medium text-slate-100">{company.has_rsu ? 'Yes' : 'No'}</span>
            </div>
            {company.has_rsu && (
              <div className="flex justify-between">
                <span className="text-slate-500">RSU Refresher</span>
                <span className="font-medium text-slate-100">
                  {company.has_rsu_refresher ? 'Yes' : 'No'}
                </span>
              </div>
            )}
          </div>
          {company.compensation_bands != null && (
            <div className="mt-3 border-t border-edge pt-3">
              <p className="text-xs text-slate-500">Compensation bands available</p>
            </div>
          )}
        </div>

        {/* Links */}
        <div className="rounded-lg border border-edge p-5">
          <h2 className="mb-3 text-sm font-semibold text-slate-100">Links</h2>
          <div className="space-y-2">
            {company.careers_page_url && (
              <a
                href={company.careers_page_url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 text-sm text-[#00f0ff] hover:text-[#00f0ff]/80"
              >
                <ExternalLink className="h-3.5 w-3.5" />
                Careers Page
              </a>
            )}
            {company.linkedin_url && (
              <a
                href={company.linkedin_url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 text-sm text-[#00f0ff] hover:text-[#00f0ff]/80"
              >
                <ExternalLink className="h-3.5 w-3.5" />
                LinkedIn
              </a>
            )}
            {company.glassdoor_url && (
              <a
                href={company.glassdoor_url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 text-sm text-[#00f0ff] hover:text-[#00f0ff]/80"
              >
                <ExternalLink className="h-3.5 w-3.5" />
                Glassdoor
              </a>
            )}
            {company.ambitionbox_url && (
              <a
                href={company.ambitionbox_url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 text-sm text-[#00f0ff] hover:text-[#00f0ff]/80"
              >
                <ExternalLink className="h-3.5 w-3.5" />
                AmbitionBox
              </a>
            )}
            {!company.careers_page_url &&
              !company.linkedin_url &&
              !company.glassdoor_url &&
              !company.ambitionbox_url && (
                <p className="text-sm text-slate-600">No links available</p>
              )}
          </div>
        </div>
      </div>

      {/* Tech Stack */}
      {company.tech_stack.length > 0 && (
        <section className="mb-8">
          <h2 className="mb-3 text-sm font-semibold text-slate-100">Tech Stack</h2>
          <TechStackTags tags={company.tech_stack} limit={0} />
        </section>
      )}

      {/* Domains */}
      {company.domains.length > 0 && (
        <section className="mb-8">
          <h2 className="mb-3 text-sm font-semibold text-slate-100">Domains</h2>
          <div className="flex flex-wrap gap-2">
            {company.domains.map((d) => (
              <span
                key={d}
                className="inline-flex items-center rounded-full bg-slate-800 px-3 py-1 text-sm text-slate-300"
              >
                {d}
              </span>
            ))}
          </div>
        </section>
      )}

      {/* Interview Patterns */}
      {company.interview_patterns != null && (
        <section className="mb-8">
          <h2 className="mb-3 text-sm font-semibold text-slate-100">Interview Process</h2>
          <div className="rounded-lg border border-edge bg-overlay p-4">
            <pre className="whitespace-pre-wrap text-sm text-slate-300">
              {typeof company.interview_patterns === 'string'
                ? company.interview_patterns
                : JSON.stringify(company.interview_patterns, null, 2)}
            </pre>
          </div>
        </section>
      )}

      {/* Your Lists — show which lists this company belongs to */}
      {isAuthenticated && uniqueLists.length > 0 && (
        <section className="mb-8">
          <h2 className="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-100">
            <List className="h-4 w-4" />
            Your Lists
          </h2>
          <div className="flex flex-wrap gap-2">
            {uniqueLists.map((l) => (
              <Link
                key={l.listId}
                href={`/lists/${l.listId}`}
                className="inline-flex items-center rounded-full border border-[#00f0ff]/30 bg-[#00f0ff]/10 px-3 py-1 text-sm font-medium text-[#00f0ff] hover:bg-[#00f0ff]/20 hover:text-[#00f0ff] transition-colors"
              >
                {l.listName}
              </Link>
            ))}
          </div>
        </section>
      )}

      {/* Your Applications — only show entries where user actually applied */}
      {isAuthenticated && appliedEntries.length > 0 && (
        <section className="mb-8">
          <h2 className="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-100">
            <Briefcase className="h-4 w-4" />
            Your Applications
          </h2>
          <div className="overflow-hidden rounded-lg border border-edge bg-card">
            <table className="min-w-full divide-y divide-edge">
              <thead className="bg-overlay">
                <tr>
                  <th className="px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                    Role
                  </th>
                  <th className="px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                    List
                  </th>
                  <th className="px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                    Tracking
                  </th>
                  <th className="px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                    App Status
                  </th>
                  <th className="hidden px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-slate-500 sm:table-cell">
                    Date Applied
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-edge">
                {appliedEntries.map((entry) => (
                  <tr key={entry.id} className="hover:bg-surface">
                    <td className="whitespace-nowrap px-4 py-2.5 text-sm font-medium text-slate-200">
                      {entry.role_title || '-'}
                    </td>
                    <td className="whitespace-nowrap px-4 py-2.5">
                      <Link
                        href={`/lists/${entry.list_id}`}
                        className="text-sm text-[#00f0ff] hover:text-[#00f0ff]/80 hover:underline"
                      >
                        {entry.list_name}
                      </Link>
                    </td>
                    <td className="whitespace-nowrap px-4 py-2.5">
                      <CompanyStatusBadge status={entry.company_status} />
                    </td>
                    <td className="whitespace-nowrap px-4 py-2.5">
                      {editingStatus === entry.id ? (
                        <select
                          value={entry.status}
                          onChange={(e) =>
                            handleStatusChange(entry, e.target.value as ApplicationStatus)
                          }
                          onBlur={() => setEditingStatus(null)}
                          autoFocus
                          className="rounded-md border border-edge-input bg-input px-2 py-1 text-xs text-slate-200 focus:border-[#00f0ff]/50 focus:outline-none"
                        >
                          {ALL_STATUSES.map((s) => (
                            <option key={s} value={s}>
                              {getStatusLabel(s)}
                            </option>
                          ))}
                        </select>
                      ) : (
                        <button onClick={() => setEditingStatus(entry.id)}>
                          <StatusBadge status={entry.status} />
                        </button>
                      )}
                    </td>
                    <td className="hidden whitespace-nowrap px-4 py-2.5 text-sm text-slate-500 sm:table-cell">
                      {entry.date_applied || '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}

      {/* Footer meta */}
      <div className="border-t border-edge pt-4 text-xs text-slate-600">
        {company.last_verified_at && (
          <span>Last verified: {new Date(company.last_verified_at).toLocaleDateString()}</span>
        )}
        {company.last_verified_at && <span className="mx-2">|</span>}
        <span>Updated: {new Date(company.updated_at).toLocaleDateString()}</span>
      </div>
    </div>
  );
}
