'use client';

import { useState, useEffect } from 'react';
import { X, Plus, Check, Loader2, Circle } from 'lucide-react';
import {
  useListsForCompany,
  useAddCompanyToList,
  useRemoveCompanyFromList,
} from '@/hooks/use-lists';

interface QuickAddToListModalProps {
  companyId: string;
  companyName: string;
  onClose: () => void;
}

export function QuickAddToListModal({
  companyId,
  companyName,
  onClose,
}: QuickAddToListModalProps) {
  const { data: lists, isLoading } = useListsForCompany(companyId);
  const addToList = useAddCompanyToList();
  const removeFromList = useRemoveCompanyFromList();
  const [pendingListId, setPendingListId] = useState<string | null>(null);

  // Close on Escape
  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', handleEsc);
    return () => document.removeEventListener('keydown', handleEsc);
  }, [onClose]);

  const handleToggle = async (listId: string, currentlyInList: boolean) => {
    setPendingListId(listId);
    try {
      if (currentlyInList) {
        await removeFromList.mutateAsync({ listId, companyId });
      } else {
        await addToList.mutateAsync({ listId, companyId });
      }
    } finally {
      setPendingListId(null);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60" onClick={onClose}>
      <div
        className="w-full max-w-sm rounded-lg border border-edge bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between border-b border-edge px-4 py-3">
          <div className="min-w-0 flex-1">
            <h3 className="text-sm font-semibold text-slate-100">Add to list</h3>
            <p className="mt-0.5 truncate text-xs text-slate-500">{companyName}</p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="ml-3 text-slate-400 hover:text-slate-200"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* List rows */}
        <div className="max-h-[300px] overflow-y-auto p-2">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="h-5 w-5 animate-spin rounded-full border-2 border-[#00f0ff] border-t-transparent" />
            </div>
          ) : !lists || lists.length === 0 ? (
            <div className="py-8 text-center text-sm text-slate-500">
              No lists yet. Create a list first.
            </div>
          ) : (
            lists.map((list) => {
              const isPending = pendingListId === list.list_id;
              const inList = list.contains_company;

              return (
                <div
                  key={list.list_id}
                  className="flex w-full items-center justify-between rounded-md px-3 py-2.5 transition-colors hover:bg-surface"
                >
                  {/* List name */}
                  <span className="truncate text-sm text-slate-200">{list.name}</span>

                  {/* Action button — far right, single-icon */}
                  <button
                    type="button"
                    onClick={() => handleToggle(list.list_id, inList)}
                    disabled={isPending}
                    className="group/btn ml-3 flex h-7 w-7 shrink-0 items-center justify-center rounded-full transition-all disabled:opacity-60"
                    title={inList ? 'Remove from list' : 'Add to list'}
                  >
                    {isPending ? (
                      <Loader2 className="h-4 w-4 animate-spin text-slate-400" />
                    ) : inList ? (
                      <>
                        {/* Default: green check. Hover: red X */}
                        <span className="flex h-6 w-6 items-center justify-center rounded-full bg-green-500/20 text-green-400 group-hover/btn:hidden">
                          <Check className="h-3.5 w-3.5" />
                        </span>
                        <span className="hidden h-6 w-6 items-center justify-center rounded-full bg-red-500/20 text-red-400 group-hover/btn:flex">
                          <X className="h-3.5 w-3.5" />
                        </span>
                      </>
                    ) : (
                      <>
                        {/* Default: empty circle. Hover: green + */}
                        <span className="flex h-6 w-6 items-center justify-center rounded-full border border-slate-700 text-slate-600 group-hover/btn:hidden">
                          <Circle className="h-3 w-3" />
                        </span>
                        <span className="hidden h-6 w-6 items-center justify-center rounded-full bg-green-500/20 text-green-400 group-hover/btn:flex">
                          <Plus className="h-3.5 w-3.5" />
                        </span>
                      </>
                    )}
                  </button>
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
}
