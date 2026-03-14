'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useLists, useCreateList, useDeleteList } from '@/hooks/use-lists';
import { useAuthStore } from '@/store/auth-store';

export default function ListsPage() {
  const { data: lists, isLoading } = useLists();
  const { isPremium } = useAuthStore();
  const createList = useCreateList();
  const deleteList = useDeleteList();

  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDesc, setNewDesc] = useState('');

  const limit = isPremium ? 5 : 3;
  const canCreate = (lists?.length ?? 0) < limit;

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName.trim()) return;
    await createList.mutateAsync({
      name: newName.trim(),
      description: newDesc.trim() || undefined,
    });
    setNewName('');
    setNewDesc('');
    setShowCreate(false);
  };

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Delete list "${name}"? All entries will be removed.`)) return;
    await deleteList.mutateAsync(id);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-[#00f0ff] border-t-transparent" />
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">My Lists</h1>
          <p className="mt-1 text-sm text-slate-500">
            {lists?.length ?? 0} of {limit} lists used
            {!isPremium && ' (upgrade for 5)'}
          </p>
        </div>
        {canCreate && (
          <button
            onClick={() => setShowCreate(true)}
            className="btn-neon rounded-md px-4 py-2 text-sm font-medium"
          >
            Create list
          </button>
        )}
      </div>

      {/* Create modal */}
      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-lg bg-card p-6 shadow-xl">
            <h2 className="text-lg font-semibold text-slate-100">Create new list</h2>
            <form onSubmit={handleCreate} className="mt-4 space-y-4">
              <div>
                <label htmlFor="list-name" className="block text-sm font-medium text-slate-300">
                  Name
                </label>
                <input
                  id="list-name"
                  type="text"
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder="e.g., Dream Companies"
                  className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 shadow-sm focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
                  autoFocus
                  maxLength={255}
                />
              </div>
              <div>
                <label htmlFor="list-desc" className="block text-sm font-medium text-slate-300">
                  Description (optional)
                </label>
                <textarea
                  id="list-desc"
                  value={newDesc}
                  onChange={(e) => setNewDesc(e.target.value)}
                  rows={2}
                  className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 shadow-sm focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
                />
              </div>
              <div className="flex justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setShowCreate(false)}
                  className="rounded-md border border-edge px-4 py-2 text-sm font-medium text-slate-300 hover:bg-overlay"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={!newName.trim() || createList.isPending}
                  className="btn-neon rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
                >
                  {createList.isPending ? 'Creating...' : 'Create'}
                </button>
              </div>
              {createList.isError && (
                <p className="text-sm text-red-400">
                  {(createList.error as Error).message}
                </p>
              )}
            </form>
          </div>
        </div>
      )}

      {/* List cards */}
      {!lists || lists.length === 0 ? (
        <div className="mt-8 rounded-lg border border-dashed border-edge p-12 text-center">
          <p className="text-sm text-slate-500">No lists yet. Create one to start tracking applications.</p>
        </div>
      ) : (
        <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {lists.map((list) => (
            <div
              key={list.id}
              className="group relative card-neon-hover rounded-lg border border-edge bg-card p-5 transition-all"
            >
              <Link href={`/lists/${list.id}`} className="block">
                <h3 className="text-lg font-semibold text-slate-100">{list.name}</h3>
                {list.description && (
                  <p className="mt-1 line-clamp-2 text-sm text-slate-500">{list.description}</p>
                )}
                <p className="mt-3 text-sm text-slate-600">
                  {list.entry_count} {list.entry_count === 1 ? 'entry' : 'entries'}
                </p>
              </Link>
              <button
                onClick={() => handleDelete(list.id, list.name)}
                className="absolute right-3 top-3 hidden rounded p-1 text-slate-600 hover:bg-red-900/30 hover:text-red-400 group-hover:block"
                title="Delete list"
              >
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
