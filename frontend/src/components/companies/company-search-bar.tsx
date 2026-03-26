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
      <Search className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-[var(--color-text-muted)]" />
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full rounded-lg border border-edge-input bg-input py-3 pl-10 pr-4 text-sm text-[var(--color-text)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30"
      />
    </div>
  );
}
