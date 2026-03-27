'use client';

import { useState, useMemo } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft, ExternalLink, MapPin, Building2, Calendar, DollarSign, Briefcase, FileText, List, PenSquare } from 'lucide-react';
import { useCompanyDetail } from '@/hooks/use-companies';
import { useApplicationsByCompany, useUpdateApplication } from '@/hooks/use-applications';
import { useAuthStore } from '@/store/auth-store';
import { TechStackTags } from '@/components/companies/tech-stack-tags';
import { StatusBadge, ALL_STATUSES, getStatusLabel } from '@/components/lists/status-badge';
import { CompanyStatusBadge } from '@/components/lists/company-status-badge';
import { ApplicationEditModal } from '@/components/applications/application-edit-modal';
import { ApplicationNotesModal } from '@/components/applications/application-notes-modal';
import type { ApplicationStatus, Application, CompanyTrackingStatus } from '@/types/api';
import { apiClient } from '@/lib/api';
import { useQuery } from '@tanstack/react-query';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import type { ListEntry } from '@/types/api';

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
  paused: 'bg-[var(--color-warning)]/15 text-[var(--color-warning)] border-[var(--color-warning)]/60',
  unknown: 'bg-slate-800 text-[var(--color-text-muted)] border-edge',
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

function useEntriesByCompany(userId: string | undefined, companyId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.companyEntries.byCompany(userId!, companyId!),
    queryFn: () =>
      apiClient.get<
        (ListEntry & { list_name: string })[]
      >('/api/entries', { company_id: companyId! }),
    enabled: !!companyId && !!userId,
    staleTime: staleTimes.userLists,
  });
}

export default function CompanyProfile() {
  const { slug } = useParams<{ slug: string }>();
  const { data: company, isLoading, isError, error } = useCompanyDetail(slug);
  const { isAuthenticated, user } = useAuthStore();
  const { data: userEntries } = useEntriesByCompany(
    isAuthenticated ? user?.id : undefined,
    isAuthenticated && company ? company.id : undefined,
  );
  const { data: applications } = useApplicationsByCompany(
    isAuthenticated && company ? company.id : undefined,
  );
  const updateApp = useUpdateApplication();
  const [editingStatus, setEditingStatus] = useState<string | null>(null);
  const [editingApplication, setEditingApplication] = useState<Application | null>(null);
  const [notesApplication, setNotesApplication] = useState<Application | null>(null);

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

  const handleStatusChange = async (app: Application, newStatus: ApplicationStatus) => {
    await updateApp.mutateAsync({ id: app.id, status: newStatus });
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
          <Link href="/companies" className="mt-4 inline-block text-sm text-[var(--color-primary)] hover:text-[var(--color-primary)]/80">
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
        className="mb-6 inline-flex items-center gap-1.5 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to directory
      </Link>

      {/* Header */}
      <div className="mb-8">
        <div className="flex items-start justify-between gap-4">
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-3xl font-bold text-[var(--color-text)]">{company.name}</h1>
              {overallStatus && <CompanyStatusBadge status={overallStatus} />}
            </div>
            <div className="mt-2 flex flex-wrap items-center gap-3 text-sm text-[var(--color-text-muted)]">
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
          <p className="text-[var(--color-text)] leading-relaxed">{company.description}</p>
        </section>
      )}

      {/* Key Info Grid */}
      <div className="mb-8 grid gap-6 sm:grid-cols-2">
        {/* Compensation */}
        <div className="rounded-lg border border-edge p-5">
          <h2 className="mb-3 flex items-center gap-2 text-sm font-semibold text-[var(--color-text)]">
            <DollarSign className="h-4 w-4" />
            Compensation
          </h2>
          <div className="space-y-2 text-sm">
            {company.compensation_tier && (
              <div className="flex justify-between">
                <span className="text-[var(--color-text-muted)]">Tier</span>
                <span className="font-medium text-[var(--color-text)]">
                  {tierLabels[company.compensation_tier] || company.compensation_tier}
                </span>
              </div>
            )}
            <div className="flex justify-between">
              <span className="text-[var(--color-text-muted)]">RSU</span>
              <span className="font-medium text-[var(--color-text)]">{company.has_rsu ? 'Yes' : 'No'}</span>
            </div>
            {company.has_rsu && (
              <div className="flex justify-between">
                <span className="text-[var(--color-text-muted)]">RSU Refresher</span>
                <span className="font-medium text-[var(--color-text)]">
                  {company.has_rsu_refresher ? 'Yes' : 'No'}
                </span>
              </div>
            )}
          </div>
          {company.compensation_bands != null && (
            <div className="mt-3 border-t border-edge pt-3">
              <p className="text-xs text-[var(--color-text-muted)]">Compensation bands available</p>
            </div>
          )}
        </div>

        {/* Links */}
        <div className="rounded-lg border border-edge p-5">
          <h2 className="mb-3 text-sm font-semibold text-[var(--color-text)]">Links</h2>
          <div className="space-y-2">
            {company.careers_page_url && (
              <a
                href={company.careers_page_url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 text-sm text-[var(--color-primary)] hover:text-[var(--color-primary)]/80"
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
                className="flex items-center gap-2 text-sm text-[var(--color-primary)] hover:text-[var(--color-primary)]/80"
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
                className="flex items-center gap-2 text-sm text-[var(--color-primary)] hover:text-[var(--color-primary)]/80"
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
                className="flex items-center gap-2 text-sm text-[var(--color-primary)] hover:text-[var(--color-primary)]/80"
              >
                <ExternalLink className="h-3.5 w-3.5" />
                AmbitionBox
              </a>
            )}
            {!company.careers_page_url &&
              !company.linkedin_url &&
              !company.glassdoor_url &&
              !company.ambitionbox_url && (
                <p className="text-sm text-[var(--color-text-muted)]">No links available</p>
              )}
          </div>
        </div>
      </div>

      {/* Tech Stack */}
      {company.tech_stack.length > 0 && (
        <section className="mb-8">
          <h2 className="mb-3 text-sm font-semibold text-[var(--color-text)]">Tech Stack</h2>
          <TechStackTags tags={company.tech_stack} limit={0} />
        </section>
      )}

      {/* Domains */}
      {company.domains.length > 0 && (
        <section className="mb-8">
          <h2 className="mb-3 text-sm font-semibold text-[var(--color-text)]">Domains</h2>
          <div className="flex flex-wrap gap-2">
            {company.domains.map((d) => (
              <span
                key={d}
                className="inline-flex items-center rounded-full bg-slate-800 px-3 py-1 text-sm text-[var(--color-text)]"
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
          <h2 className="mb-3 text-sm font-semibold text-[var(--color-text)]">Interview Process</h2>
          <div className="rounded-lg border border-edge bg-overlay p-4">
            <pre className="whitespace-pre-wrap text-sm text-[var(--color-text)]">
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
          <h2 className="mb-3 flex items-center gap-2 text-sm font-semibold text-[var(--color-text)]">
            <List className="h-4 w-4" />
            Your Lists
          </h2>
          <div className="flex flex-wrap gap-2">
            {uniqueLists.map((l) => (
              <Link
                key={l.listId}
                href={`/lists/${l.listId}`}
                className="inline-flex items-center rounded-full border border-[var(--color-primary)]/30 bg-[var(--color-primary)]/10 px-3 py-1 text-sm font-medium text-[var(--color-primary)] hover:bg-[var(--color-primary)]/20 hover:text-[var(--color-primary)] transition-colors"
              >
                {l.listName}
              </Link>
            ))}
          </div>
        </section>
      )}

      {/* Your Applications — from the applications table */}
      {isAuthenticated && applications && applications.length > 0 && (
        <section className="mb-8">
          <h2 className="mb-3 flex items-center gap-2 text-sm font-semibold text-[var(--color-text)]">
            <Briefcase className="h-4 w-4" />
            Your Applications
          </h2>
          <div className="overflow-x-auto rounded-lg border border-edge bg-card">
            <table className="min-w-full divide-y divide-edge">
              <thead className="bg-overlay">
                <tr>
                  <th className="px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">
                    Role
                  </th>
                  <th className="px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">
                    Status
                  </th>
                  <th className="hidden px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)] sm:table-cell">
                    Date Applied
                  </th>
                  <th className="px-4 py-2 text-right text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-edge">
                {applications.map((app) => (
                  <tr key={app.id} className="hover:bg-surface">
                    <td className="whitespace-nowrap px-4 py-2.5 text-sm font-medium text-[var(--color-text)]">
                      {app.role_title || '-'}
                    </td>
                    <td className="whitespace-nowrap px-4 py-2.5">
                      {editingStatus === app.id ? (
                        <select
                          value={app.status}
                          onChange={(e) =>
                            handleStatusChange(app, e.target.value as ApplicationStatus)
                          }
                          onBlur={() => setEditingStatus(null)}
                          autoFocus
                          className="rounded-md border border-edge-input bg-input px-2 py-1 text-xs text-[var(--color-text)] focus:border-[var(--color-primary)]/50 focus:outline-none"
                        >
                          {ALL_STATUSES.map((s) => (
                            <option key={s} value={s}>
                              {getStatusLabel(s)}
                            </option>
                          ))}
                        </select>
                      ) : (
                        <button onClick={() => setEditingStatus(app.id)}>
                          <StatusBadge status={app.status} />
                        </button>
                      )}
                    </td>
                    <td className="hidden whitespace-nowrap px-4 py-2.5 text-sm text-[var(--color-text-muted)] sm:table-cell">
                      {app.date_applied || '-'}
                    </td>
                    <td className="px-4 py-2.5">
                      <div className="flex justify-end gap-2">
                        {app.notes && (
                          <button
                            type="button"
                            onClick={() => setNotesApplication(app)}
                            className="inline-flex items-center gap-1 rounded-md border border-edge px-2 py-1 text-xs font-medium text-[var(--color-text-muted)] transition-colors hover:bg-overlay hover:text-[var(--color-text)]"
                          >
                            <FileText className="h-3.5 w-3.5" />
                            Notes
                          </button>
                        )}
                        <button
                          type="button"
                          onClick={() => setEditingApplication(app)}
                          className="inline-flex items-center gap-1 rounded-md border border-edge px-2 py-1 text-xs font-medium text-[var(--color-text-muted)] transition-colors hover:bg-overlay hover:text-[var(--color-text)]"
                        >
                          <PenSquare className="h-3.5 w-3.5" />
                          Edit
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}

      {/* Footer meta */}
      <div className="border-t border-edge pt-4 text-xs text-[var(--color-text-muted)]">
        {company.last_verified_at && (
          <span>Last verified: {new Date(company.last_verified_at).toLocaleDateString()}</span>
        )}
        {company.last_verified_at && <span className="mx-2">|</span>}
        <span>Updated: {new Date(company.updated_at).toLocaleDateString()}</span>
      </div>

      {editingApplication && (
        <ApplicationEditModal
          application={editingApplication}
          companyName={company.name}
          onClose={() => setEditingApplication(null)}
        />
      )}

      {notesApplication && (
        <ApplicationNotesModal
          application={notesApplication}
          companyName={company.name}
          onClose={() => setNotesApplication(null)}
        />
      )}
    </div>
  );
}
