'use client';

import { Search } from 'lucide-react';

interface CompanySearchBarProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export function CompanySearchBar({
  value,
  onChange,
  placeholder = 'Search companies, tech stacks, domains...',
}: CompanySearchBarProps) {
  return (
    <div className="relative">
      <Search className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-slate-500" />
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full rounded-lg border border-edge-input bg-input py-3 pl-10 pr-4 text-sm text-slate-200 focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
      />
    </div>
  );
}
