'use client';

import { useState, useEffect } from 'react';
import { X } from 'lucide-react';
import { useAdminCreateCompany, useAdminUpdateCompany, useAdminUploadLogo } from '@/hooks/use-admin';
import type { CompanyDetail } from '@/types/api';

interface CompanyModalProps {
  company?: CompanyDetail | null;
  onClose: () => void;
  onSuccess?: (message: string) => void;
}

export function CompanyModal({ company, onClose, onSuccess }: CompanyModalProps) {
  const isEdit = !!company;
  const createCompany = useAdminCreateCompany();
  const updateCompany = useAdminUpdateCompany();
  const uploadLogo = useAdminUploadLogo();

  const [form, setForm] = useState({
    name: company?.name || '',
    slug: company?.slug || '',
    description: company?.description || '',
    size: company?.size || '',
    headquarters: company?.headquarters || '',
    founded_year: company?.founded_year?.toString() || '',
    careers_page_url: company?.careers_page_url || '',
    glassdoor_url: company?.glassdoor_url || '',
    ambitionbox_url: company?.ambitionbox_url || '',
    linkedin_url: company?.linkedin_url || '',
    hiring_status: company?.hiring_status || 'unknown',
    compensation_tier: company?.compensation_tier || '',
    has_rsu: company?.has_rsu || false,
    has_rsu_refresher: company?.has_rsu_refresher || false,
    tech_stack: company?.tech_stack?.join(', ') || '',
    domains: company?.domains?.join(', ') || '',
    office_modes: company?.office_modes?.join(', ') || '',
  });
  const [logoFile, setLogoFile] = useState<File | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [onClose]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    const splitTrim = (s: string) =>
      s
        .split(',')
        .map((v) => v.trim())
        .filter(Boolean);

    const payload: Record<string, unknown> = {
      name: form.name,
      slug: form.slug,
      description: form.description || null,
      size: form.size || null,
      headquarters: form.headquarters || null,
      founded_year: form.founded_year ? parseInt(form.founded_year) : null,
      careers_page_url: form.careers_page_url || null,
      glassdoor_url: form.glassdoor_url || null,
      ambitionbox_url: form.ambitionbox_url || null,
      linkedin_url: form.linkedin_url || null,
      hiring_status: form.hiring_status,
      compensation_tier: form.compensation_tier || null,
      has_rsu: form.has_rsu,
      has_rsu_refresher: form.has_rsu_refresher,
      tech_stack: splitTrim(form.tech_stack),
      domains: splitTrim(form.domains),
      office_modes: splitTrim(form.office_modes),
    };

    try {
      let companyId = company?.id;

      if (isEdit && companyId) {
        await updateCompany.mutateAsync({ companyId, ...payload });
      } else {
        const created = await createCompany.mutateAsync(payload);
        companyId = created.id;
      }

      if (logoFile && companyId) {
        const fd = new FormData();
        fd.append('logo', logoFile);
        await uploadLogo.mutateAsync({ companyId, formData: fd });
      }

      onSuccess?.(isEdit ? 'Company updated successfully.' : 'Company created successfully.');
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    }
  };

  const isPending =
    createCompany.isPending || updateCompany.isPending || uploadLogo.isPending;

  const inputClass =
    'mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 placeholder:text-slate-600 focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30';
  const labelClass = 'block text-sm font-medium text-slate-300';

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/60 py-8"
      onClick={onClose}
    >
      <div
        className="w-full max-w-2xl rounded-lg border border-edge bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b border-edge px-5 py-4">
          <h3 className="text-sm font-semibold text-slate-100">
            {isEdit ? 'Edit Company' : 'Create Company'}
          </h3>
          <button type="button" onClick={onClose} className="text-slate-400 hover:text-slate-200">
            <X className="h-5 w-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 px-5 py-4">
          {error && (
            <div className="rounded-md border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-400">
              {error}
            </div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="name" className={labelClass}>Name *</label>
              <input id="name" required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} className={inputClass} />
            </div>
            <div>
              <label htmlFor="slug" className={labelClass}>Slug *</label>
              <input id="slug" required value={form.slug} onChange={(e) => setForm({ ...form, slug: e.target.value })} className={inputClass} placeholder="e.g. google" />
            </div>
          </div>

          <div>
            <label htmlFor="description" className={labelClass}>Description</label>
            <textarea id="description" rows={2} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} className={inputClass} />
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div>
              <label htmlFor="size" className={labelClass}>Size</label>
              <select id="size" value={form.size} onChange={(e) => setForm({ ...form, size: e.target.value })} className={inputClass}>
                <option value="">-</option>
                <option value="startup">Startup</option>
                <option value="small">Small</option>
                <option value="mid">Mid</option>
                <option value="large">Large</option>
                <option value="enterprise">Enterprise</option>
              </select>
            </div>
            <div>
              <label htmlFor="headquarters" className={labelClass}>Headquarters</label>
              <input id="headquarters" value={form.headquarters} onChange={(e) => setForm({ ...form, headquarters: e.target.value })} className={inputClass} />
            </div>
            <div>
              <label htmlFor="founded_year" className={labelClass}>Founded Year</label>
              <input id="founded_year" type="number" value={form.founded_year} onChange={(e) => setForm({ ...form, founded_year: e.target.value })} className={inputClass} />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="hiring_status" className={labelClass}>Hiring Status</label>
              <select id="hiring_status" value={form.hiring_status} onChange={(e) => setForm({ ...form, hiring_status: e.target.value })} className={inputClass}>
                <option value="active">Active</option>
                <option value="paused">Paused</option>
                <option value="unknown">Unknown</option>
              </select>
            </div>
            <div>
              <label htmlFor="compensation_tier" className={labelClass}>Compensation Tier</label>
              <input id="compensation_tier" value={form.compensation_tier} onChange={(e) => setForm({ ...form, compensation_tier: e.target.value })} className={inputClass} placeholder="e.g. A, B, C" />
            </div>
          </div>

          <div className="flex items-center gap-6">
            <label className="flex items-center gap-2 text-sm text-slate-300">
              <input type="checkbox" checked={form.has_rsu} onChange={(e) => setForm({ ...form, has_rsu: e.target.checked })} className="rounded border-slate-600" />
              Has RSU
            </label>
            <label className="flex items-center gap-2 text-sm text-slate-300">
              <input type="checkbox" checked={form.has_rsu_refresher} onChange={(e) => setForm({ ...form, has_rsu_refresher: e.target.checked })} className="rounded border-slate-600" />
              RSU Refresher
            </label>
          </div>

          <div>
            <label htmlFor="tech_stack" className={labelClass}>Tech Stack (comma-separated)</label>
            <input id="tech_stack" value={form.tech_stack} onChange={(e) => setForm({ ...form, tech_stack: e.target.value })} className={inputClass} placeholder="React, Go, PostgreSQL" />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="domains" className={labelClass}>Domains (comma-separated)</label>
              <input id="domains" value={form.domains} onChange={(e) => setForm({ ...form, domains: e.target.value })} className={inputClass} placeholder="fintech, e-commerce" />
            </div>
            <div>
              <label htmlFor="office_modes" className={labelClass}>Office Modes (comma-separated)</label>
              <input id="office_modes" value={form.office_modes} onChange={(e) => setForm({ ...form, office_modes: e.target.value })} className={inputClass} placeholder="remote, hybrid, onsite" />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="careers_url" className={labelClass}>Careers Page URL</label>
              <input id="careers_url" value={form.careers_page_url} onChange={(e) => setForm({ ...form, careers_page_url: e.target.value })} className={inputClass} />
            </div>
            <div>
              <label htmlFor="linkedin_url" className={labelClass}>LinkedIn URL</label>
              <input id="linkedin_url" value={form.linkedin_url} onChange={(e) => setForm({ ...form, linkedin_url: e.target.value })} className={inputClass} />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="glassdoor_url" className={labelClass}>Glassdoor URL</label>
              <input id="glassdoor_url" value={form.glassdoor_url} onChange={(e) => setForm({ ...form, glassdoor_url: e.target.value })} className={inputClass} />
            </div>
            <div>
              <label htmlFor="ambitionbox_url" className={labelClass}>AmbitionBox URL</label>
              <input id="ambitionbox_url" value={form.ambitionbox_url} onChange={(e) => setForm({ ...form, ambitionbox_url: e.target.value })} className={inputClass} />
            </div>
          </div>

          <div>
            <label htmlFor="logo" className={labelClass}>Logo (image, max 2MB)</label>
            <input
              id="logo"
              type="file"
              accept="image/*"
              onChange={(e) => setLogoFile(e.target.files?.[0] || null)}
              className="mt-1 block w-full text-sm text-slate-400 file:mr-4 file:rounded-md file:border-0 file:bg-[#00f0ff]/10 file:px-3 file:py-1.5 file:text-sm file:text-[#00f0ff] hover:file:bg-[#00f0ff]/20"
            />
          </div>

          <div className="flex justify-end gap-3 border-t border-edge pt-4">
            <button
              type="button"
              onClick={onClose}
              className="rounded-md border border-edge px-4 py-2 text-sm font-medium text-slate-300 hover:bg-card hover:text-slate-100"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isPending}
              className="btn-neon rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
            >
              {isPending ? 'Saving...' : isEdit ? 'Update Company' : 'Create Company'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
