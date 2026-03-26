'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import { Search, X, Building2 } from 'lucide-react';
import { useCompanySearch } from '@/hooks/use-companies';

interface CompanySelection {
  id: string;
  name: string;
}

interface CompanyComboboxProps {
  value: CompanySelection | null;
  onChange: (company: CompanySelection | null) => void;
  placeholder?: string;
}

export function CompanyCombobox({
  value,
  onChange,
  placeholder = 'Search companies...',
}: CompanyComboboxProps) {
  const [query, setQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const [isOpen, setIsOpen] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const { data: results, isLoading } = useCompanySearch(debouncedQuery);

  // Debounce search query
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(query), 300);
    return () => clearTimeout(timer);
  }, [query]);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleSelect = useCallback(
    (company: { id: string; name: string }) => {
      onChange({ id: company.id, name: company.name });
      setQuery('');
      setIsOpen(false);
    },
    [onChange],
  );

  const handleClear = useCallback(() => {
    onChange(null);
    setQuery('');
    setIsOpen(false);
    inputRef.current?.focus();
  }, [onChange]);

  // If a company is selected, show its name
  if (value) {
    return (
      <div className="flex items-center gap-2 rounded-md border border-edge-input bg-input px-3 py-2">
        <Building2 className="h-4 w-4 shrink-0 text-[var(--color-text-muted)]" />
        <span className="flex-1 truncate text-sm text-[var(--color-text)]">{value.name}</span>
        <button
          type="button"
          onClick={handleClear}
          className="shrink-0 text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
        >
          <X className="h-4 w-4" />
        </button>
      </div>
    );
  }

  return (
    <div ref={wrapperRef} className="relative">
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--color-text-muted)]" />
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setIsOpen(true);
          }}
          onFocus={() => {
            if (query.length >= 2) setIsOpen(true);
          }}
          placeholder={placeholder}
          className="block w-full rounded-md border border-edge-input bg-input py-2 pl-9 pr-3 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30"
        />
      </div>

      {isOpen && debouncedQuery.length >= 2 && (
        <div className="absolute z-50 mt-1 max-h-56 w-full overflow-auto rounded-md border border-edge bg-card shadow-lg shadow-black/30">
          {isLoading ? (
            <div className="px-4 py-3 text-sm text-[var(--color-text-muted)]">Searching...</div>
          ) : results && results.length > 0 ? (
            results.map((company) => (
              <button
                key={company.id}
                type="button"
                onClick={() => handleSelect(company)}
                className="flex w-full items-center gap-3 px-4 py-2.5 text-left text-sm hover:bg-surface"
              >
                <Building2 className="h-4 w-4 shrink-0 text-[var(--color-text-muted)]" />
                <div className="min-w-0 flex-1">
                  <div className="truncate font-medium text-[var(--color-text)]">{company.name}</div>
                  <div className="truncate text-xs text-[var(--color-text-muted)]">
                    {[company.headquarters, company.size].filter(Boolean).join(' · ')}
                  </div>
                </div>
                {company.hiring_status === 'active' && (
                  <span className="shrink-0 rounded-full bg-green-900/30 px-2 py-0.5 text-[10px] text-green-400">
                    Hiring
                  </span>
                )}
              </button>
            ))
          ) : (
            <div className="px-4 py-3 text-sm text-[var(--color-text-muted)]">No companies found</div>
          )}
        </div>
      )}
    </div>
  );
}
