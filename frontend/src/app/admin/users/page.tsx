'use client';

import { useState, useEffect } from 'react';
import { Search, Coins, Shield, ShieldOff, Crown } from 'lucide-react';
import { useAdminUsers, useAdminUpdateUser } from '@/hooks/use-admin';
import { CreditModal } from '@/components/admin/credit-modal';
import type { AdminUser } from '@/types/api';

export default function AdminUsersPage() {
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [roleFilter, setRoleFilter] = useState('');
  const [creditTarget, setCreditTarget] = useState<AdminUser | null>(null);
  const updateUser = useAdminUpdateUser();

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search), 300);
    return () => clearTimeout(timer);
  }, [search]);

  const { data, isLoading } = useAdminUsers({
    q: debouncedSearch || undefined,
    role: roleFilter || undefined,
    limit: '50',
  });

  const users = data?.data ?? [];
  const total = data?.total ?? 0;

  const handleBanToggle = async (user: AdminUser) => {
    const isBanned = !!user.deleted_at;
    if (!isBanned && !confirm(`Ban user "${user.name}" (${user.email})?`)) return;
    await updateUser.mutateAsync({ userId: user.id, banned: !isBanned });
  };

  const handlePremiumToggle = async (user: AdminUser) => {
    const isPremium = !!user.premium_since;
    await updateUser.mutateAsync({
      userId: user.id,
      set_premium: !isPremium,
    });
  };

  const handleRoleChange = async (user: AdminUser, newRole: string) => {
    if (newRole === user.role) return;
    await updateUser.mutateAsync({ userId: user.id, role: newRole });
  };

  const inputClass =
    'block w-full rounded-md border border-edge-input bg-input py-2 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30';

  return (
    <div>
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold text-[var(--color-text)]">Users</h1>
        <span className="text-sm text-[var(--color-text-muted)]">{total} total</span>
      </div>

      <div className="mt-4 flex gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--color-text-muted)]" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search by name or email..."
            className={`${inputClass} pl-10 pr-3`}
          />
        </div>
        <select
          value={roleFilter}
          onChange={(e) => setRoleFilter(e.target.value)}
          className={`${inputClass} w-36 px-3`}
        >
          <option value="">All roles</option>
          <option value="user">User</option>
          <option value="moderator">Moderator</option>
          <option value="admin">Admin</option>
        </select>
      </div>

      <div className="mt-4 overflow-x-auto rounded-lg border border-edge">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-edge bg-overlay text-left">
              <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Name</th>
              <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Email</th>
              <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Role</th>
              <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Premium</th>
              <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Status</th>
              <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Joined</th>
              <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-[var(--color-text-muted)]">
                  <div className="flex items-center justify-center">
                    <div className="h-5 w-5 animate-spin rounded-full border-2 border-[var(--color-primary)] border-t-transparent" />
                  </div>
                </td>
              </tr>
            ) : users.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-[var(--color-text-muted)]">
                  No users found.
                </td>
              </tr>
            ) : (
              users.map((u) => {
                const isBanned = !!u.deleted_at;
                const isPremium = !!u.premium_since;
                return (
                  <tr key={u.id} className={`border-b border-edge ${isBanned ? 'opacity-50' : 'hover:bg-card/50'}`}>
                    <td className="px-4 py-3 font-medium text-[var(--color-text)]">{u.name}</td>
                    <td className="px-4 py-3 text-[var(--color-text-muted)]">{u.email}</td>
                    <td className="px-4 py-3">
                      <select
                        value={u.role}
                        onChange={(e) => handleRoleChange(u, e.target.value)}
                        disabled={updateUser.isPending}
                        className="rounded border border-edge bg-input px-2 py-1 text-xs text-[var(--color-text)]"
                      >
                        <option value="user">user</option>
                        <option value="moderator">moderator</option>
                        <option value="admin">admin</option>
                      </select>
                    </td>
                    <td className="px-4 py-3">
                      {isPremium ? (
                        <span className="inline-flex items-center gap-1 rounded-full bg-[var(--color-warning)]/15 px-2 py-0.5 text-xs font-medium text-[var(--color-warning)]">
                          <Crown className="h-3 w-3" /> Premium
                        </span>
                      ) : (
                        <span className="text-xs text-[var(--color-text-muted)]">Free</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {isBanned ? (
                        <span className="inline-block rounded-full bg-red-500/15 px-2 py-0.5 text-xs font-medium text-red-400">
                          Banned
                        </span>
                      ) : (
                        <span className="inline-block rounded-full bg-green-500/15 px-2 py-0.5 text-xs font-medium text-green-400">
                          Active
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-[var(--color-text-muted)]">
                      {new Date(u.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1">
                        <button
                          onClick={() => handleBanToggle(u)}
                          disabled={updateUser.isPending}
                          title={isBanned ? 'Unban' : 'Ban'}
                          className={`rounded-md p-1.5 ${
                            isBanned
                              ? 'text-green-400 hover:bg-green-500/10'
                              : 'text-red-400 hover:bg-red-500/10'
                          }`}
                        >
                          {isBanned ? <ShieldOff className="h-4 w-4" /> : <Shield className="h-4 w-4" />}
                        </button>
                        <button
                          onClick={() => handlePremiumToggle(u)}
                          disabled={updateUser.isPending}
                          title={isPremium ? 'Revoke premium' : 'Grant premium'}
                          className={`rounded-md p-1.5 ${
                            isPremium
                              ? 'text-[var(--color-warning)] hover:bg-[var(--color-warning)]/10'
                              : 'text-[var(--color-text-muted)] hover:bg-card'
                          }`}
                        >
                          <Crown className="h-4 w-4" />
                        </button>
                        <button
                          onClick={() => setCreditTarget(u)}
                          title="Allocate credits"
                          className="rounded-md p-1.5 text-[var(--color-primary)] hover:bg-[var(--color-primary)]/10"
                        >
                          <Coins className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {creditTarget && (
        <CreditModal
          userId={creditTarget.id}
          userName={`${creditTarget.name} (${creditTarget.email})`}
          onClose={() => setCreditTarget(null)}
        />
      )}
    </div>
  );
}
