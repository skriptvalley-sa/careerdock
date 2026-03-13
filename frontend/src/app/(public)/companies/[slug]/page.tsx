'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft, ExternalLink, MapPin, Building2, Calendar, DollarSign } from 'lucide-react';
import { useCompanyDetail } from '@/hooks/use-companies';
import { TechStackTags } from '@/components/companies/tech-stack-tags';

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
  active: 'bg-green-50 text-green-700 border-green-200',
  paused: 'bg-yellow-50 text-yellow-700 border-yellow-200',
  unknown: 'bg-gray-50 text-gray-600 border-gray-200',
};

export default function CompanyProfilePage() {
  const { slug } = useParams<{ slug: string }>();
  const { data: company, isLoading, isError, error } = useCompanyDetail(slug);

  if (isLoading) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-12 sm:px-6 lg:px-8">
        <div className="animate-pulse space-y-6">
          <div className="h-8 w-64 rounded bg-gray-200" />
          <div className="h-4 w-96 rounded bg-gray-200" />
          <div className="h-48 rounded-lg bg-gray-200" />
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-12 sm:px-6 lg:px-8">
        <div className="rounded-lg bg-red-50 p-6 text-center">
          <p className="text-sm text-red-700">
            Failed to load company: {(error as Error).message}
          </p>
          <Link href="/companies" className="mt-4 inline-block text-sm text-blue-600 hover:text-blue-700">
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
        className="mb-6 inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to directory
      </Link>

      {/* Header */}
      <div className="mb-8">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{company.name}</h1>
            <div className="mt-2 flex flex-wrap items-center gap-3 text-sm text-gray-500">
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
          <p className="text-gray-700 leading-relaxed">{company.description}</p>
        </section>
      )}

      {/* Key Info Grid */}
      <div className="mb-8 grid gap-6 sm:grid-cols-2">
        {/* Compensation */}
        <div className="rounded-lg border border-gray-200 p-5">
          <h2 className="mb-3 flex items-center gap-2 text-sm font-semibold text-gray-900">
            <DollarSign className="h-4 w-4" />
            Compensation
          </h2>
          <div className="space-y-2 text-sm">
            {company.compensation_tier && (
              <div className="flex justify-between">
                <span className="text-gray-500">Tier</span>
                <span className="font-medium text-gray-900">
                  {tierLabels[company.compensation_tier] || company.compensation_tier}
                </span>
              </div>
            )}
            <div className="flex justify-between">
              <span className="text-gray-500">RSU</span>
              <span className="font-medium text-gray-900">{company.has_rsu ? 'Yes' : 'No'}</span>
            </div>
            {company.has_rsu && (
              <div className="flex justify-between">
                <span className="text-gray-500">RSU Refresher</span>
                <span className="font-medium text-gray-900">
                  {company.has_rsu_refresher ? 'Yes' : 'No'}
                </span>
              </div>
            )}
          </div>
          {company.compensation_bands != null && (
            <div className="mt-3 border-t border-gray-100 pt-3">
              <p className="text-xs text-gray-500">Compensation bands available</p>
            </div>
          )}
        </div>

        {/* Links */}
        <div className="rounded-lg border border-gray-200 p-5">
          <h2 className="mb-3 text-sm font-semibold text-gray-900">Links</h2>
          <div className="space-y-2">
            {company.careers_page_url && (
              <a
                href={company.careers_page_url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 text-sm text-blue-600 hover:text-blue-700"
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
                className="flex items-center gap-2 text-sm text-blue-600 hover:text-blue-700"
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
                className="flex items-center gap-2 text-sm text-blue-600 hover:text-blue-700"
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
                className="flex items-center gap-2 text-sm text-blue-600 hover:text-blue-700"
              >
                <ExternalLink className="h-3.5 w-3.5" />
                AmbitionBox
              </a>
            )}
            {!company.careers_page_url &&
              !company.linkedin_url &&
              !company.glassdoor_url &&
              !company.ambitionbox_url && (
                <p className="text-sm text-gray-400">No links available</p>
              )}
          </div>
        </div>
      </div>

      {/* Tech Stack */}
      {company.tech_stack.length > 0 && (
        <section className="mb-8">
          <h2 className="mb-3 text-sm font-semibold text-gray-900">Tech Stack</h2>
          <TechStackTags tags={company.tech_stack} limit={0} />
        </section>
      )}

      {/* Domains */}
      {company.domains.length > 0 && (
        <section className="mb-8">
          <h2 className="mb-3 text-sm font-semibold text-gray-900">Domains</h2>
          <div className="flex flex-wrap gap-2">
            {company.domains.map((d) => (
              <span
                key={d}
                className="inline-flex items-center rounded-full bg-gray-100 px-3 py-1 text-sm text-gray-700"
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
          <h2 className="mb-3 text-sm font-semibold text-gray-900">Interview Process</h2>
          <div className="rounded-lg border border-gray-200 bg-gray-50 p-4">
            <pre className="whitespace-pre-wrap text-sm text-gray-700">
              {typeof company.interview_patterns === 'string'
                ? company.interview_patterns
                : JSON.stringify(company.interview_patterns, null, 2)}
            </pre>
          </div>
        </section>
      )}

      {/* Footer meta */}
      <div className="border-t border-gray-200 pt-4 text-xs text-gray-400">
        {company.last_verified_at && (
          <span>Last verified: {new Date(company.last_verified_at).toLocaleDateString()}</span>
        )}
        {company.last_verified_at && <span className="mx-2">|</span>}
        <span>Updated: {new Date(company.updated_at).toLocaleDateString()}</span>
      </div>
    </div>
  );
}
