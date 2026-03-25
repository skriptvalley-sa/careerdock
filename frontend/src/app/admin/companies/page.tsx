'use client';

import { useState, useEffect } from 'react';
import { Plus, Pencil, Search } from 'lucide-react';
import { useCompanyList } from '@/hooks/use-companies';
import { CompanyModal } from '@/components/admin/company-modal';
import type { CompanyDetail, CompanyListItem } from '@/types/api';
import { apiClient } from '@/lib/api';

export default function AdminCompaniesPage() {
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [modalOpen, setModalOpen] = useState(false);
  const [editCompany, setEditCompany] = useState<CompanyDetail | null>(null);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search), 300);
    return () => clearTimeout(timer);
  }, [search]);

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
      // If slug fetch fails, use list item as partial data
      setEditCompany(c as unknown as CompanyDetail);
      setModalOpen(true);
    }
  };

  const handleCreate = () => {
    setEditCompany(null);
    setModalOpen(true);
  };

  return (
    <div>
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold text-slate-100">Companies</h1>
        <button onClick={handleCreate} className="btn-neon flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium">
          <Plus className="h-4 w-4" />
          Create
        </button>
      </div>

      <div className="relative mt-4">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search companies..."
          className="block w-full rounded-md border border-edge-input bg-input py-2 pl-10 pr-3 text-sm text-slate-200 placeholder:text-slate-600 focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
        />
      </div>

      <div className="mt-4 overflow-x-auto rounded-lg border border-edge">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-edge bg-overlay text-left">
              <th className="px-4 py-3 font-medium text-slate-400">Name</th>
              <th className="px-4 py-3 font-medium text-slate-400">Slug</th>
              <th className="px-4 py-3 font-medium text-slate-400">Size</th>
              <th className="px-4 py-3 font-medium text-slate-400">HQ</th>
              <th className="px-4 py-3 font-medium text-slate-400">Hiring</th>
              <th className="px-4 py-3 font-medium text-slate-400">Tech Stack</th>
              <th className="px-4 py-3 font-medium text-slate-400" />
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-slate-500">
                  <div className="flex items-center justify-center">
                    <div className="h-5 w-5 animate-spin rounded-full border-2 border-[#00f0ff] border-t-transparent" />
                  </div>
                </td>
              </tr>
            ) : companies.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-slate-500">
                  No companies found.
                </td>
              </tr>
            ) : (
              companies.map((c) => (
                <tr key={c.id} className="border-b border-edge hover:bg-card/50">
                  <td className="px-4 py-3 font-medium text-slate-200">{c.name}</td>
                  <td className="px-4 py-3 text-slate-400">{c.slug}</td>
                  <td className="px-4 py-3 text-slate-400">{c.size || '-'}</td>
                  <td className="px-4 py-3 text-slate-400">{c.headquarters || '-'}</td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                        c.hiring_status === 'active'
                          ? 'bg-green-500/15 text-green-400'
                          : c.hiring_status === 'paused'
                            ? 'bg-amber-500/15 text-amber-400'
                            : 'bg-slate-500/15 text-slate-400'
                      }`}
                    >
                      {c.hiring_status}
                    </span>
                  </td>
                  <td className="max-w-[200px] truncate px-4 py-3 text-slate-500">
                    {c.tech_stack.slice(0, 3).join(', ')}
                    {c.tech_stack.length > 3 && ` +${c.tech_stack.length - 3}`}
                  </td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => handleEdit(c)}
                      className="rounded-md p-1.5 text-slate-400 hover:bg-card hover:text-slate-200"
                      title="Edit"
                    >
                      <Pencil className="h-4 w-4" />
                    </button>
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
        />
      )}
    </div>
  );
}
