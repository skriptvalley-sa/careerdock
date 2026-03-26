'use client';

import { useState } from 'react';
import { useAdminPayments, useAdminCreditTransactions } from '@/hooks/use-admin';

type Tab = 'payments' | 'credits';

export default function AdminPaymentsPage() {
  const [tab, setTab] = useState<Tab>('payments');
  const [userIdFilter, setUserIdFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [creditTypeFilter, setCreditTypeFilter] = useState('');

  const { data: paymentData, isLoading: paymentsLoading } = useAdminPayments({
    user_id: userIdFilter || undefined,
    status: statusFilter || undefined,
    limit: '50',
  });

  const { data: txnData, isLoading: txnsLoading } = useAdminCreditTransactions({
    user_id: userIdFilter || undefined,
    credit_type: creditTypeFilter || undefined,
    limit: '50',
  });

  const payments = paymentData?.data ?? [];
  const txns = txnData?.data ?? [];

  const inputClass =
    'block rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30';

  const formatAmount = (paise: number) => `\u20B9${(paise / 100).toFixed(2)}`;

  return (
    <div>
      <h1 className="text-xl font-bold text-[var(--color-text)]">Payments & Credits</h1>

      {/* Tab bar */}
      <div className="mt-4 flex gap-1 rounded-lg border border-edge bg-overlay p-1">
        <button
          onClick={() => setTab('payments')}
          className={`flex-1 rounded-md px-4 py-2 text-sm font-medium transition-colors ${
            tab === 'payments'
              ? 'bg-[var(--color-primary)]/10 text-[var(--color-primary)]'
              : 'text-[var(--color-text-muted)] hover:text-[var(--color-text)]'
          }`}
        >
          Payments ({paymentData?.total ?? 0})
        </button>
        <button
          onClick={() => setTab('credits')}
          className={`flex-1 rounded-md px-4 py-2 text-sm font-medium transition-colors ${
            tab === 'credits'
              ? 'bg-[var(--color-primary)]/10 text-[var(--color-primary)]'
              : 'text-[var(--color-text-muted)] hover:text-[var(--color-text)]'
          }`}
        >
          Credit Transactions ({txnData?.total ?? 0})
        </button>
      </div>

      {/* Filters */}
      <div className="mt-4 flex gap-3">
        <input
          type="text"
          value={userIdFilter}
          onChange={(e) => setUserIdFilter(e.target.value)}
          placeholder="Filter by user ID..."
          className={`${inputClass} flex-1`}
        />
        {tab === 'payments' && (
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className={`${inputClass} w-36`}
          >
            <option value="">All statuses</option>
            <option value="created">Created</option>
            <option value="captured">Captured</option>
            <option value="failed">Failed</option>
            <option value="refunded">Refunded</option>
          </select>
        )}
        {tab === 'credits' && (
          <select
            value={creditTypeFilter}
            onChange={(e) => setCreditTypeFilter(e.target.value)}
            className={`${inputClass} w-40`}
          >
            <option value="">All types</option>
            <option value="resume_upload">Resume Upload</option>
            <option value="ats_check">ATS Check</option>
            <option value="curated_list">Curated List</option>
            <option value="cv_generation">CV Generation</option>
          </select>
        )}
      </div>

      {/* Payments table */}
      {tab === 'payments' && (
        <div className="mt-4 overflow-x-auto rounded-lg border border-edge">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-edge bg-overlay text-left">
                <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">User ID</th>
                <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Product</th>
                <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Amount</th>
                <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Status</th>
                <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Razorpay Order</th>
                <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Date</th>
              </tr>
            </thead>
            <tbody>
              {paymentsLoading ? (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-[var(--color-text-muted)]">
                    <div className="flex items-center justify-center">
                      <div className="h-5 w-5 animate-spin rounded-full border-2 border-[var(--color-primary)] border-t-transparent" />
                    </div>
                  </td>
                </tr>
              ) : payments.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-[var(--color-text-muted)]">
                    No payments found.
                  </td>
                </tr>
              ) : (
                payments.map((p) => (
                  <tr key={p.id} className="border-b border-edge hover:bg-card/50">
                    <td className="px-4 py-3 font-mono text-xs text-[var(--color-text-muted)]">
                      {p.user_id.slice(0, 8)}...
                    </td>
                    <td className="px-4 py-3 text-[var(--color-text)]">{p.product_type}</td>
                    <td className="px-4 py-3 text-[var(--color-text)]">{formatAmount(p.amount_paise)}</td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                          p.status === 'captured'
                            ? 'bg-green-500/15 text-green-400'
                            : p.status === 'failed'
                              ? 'bg-red-500/15 text-red-400'
                              : p.status === 'refunded'
                                ? 'bg-[var(--color-primary)]/15 text-[var(--color-primary)]'
                                : 'bg-slate-500/15 text-[var(--color-text-muted)]'
                        }`}
                      >
                        {p.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-[var(--color-text-muted)]">
                      {p.razorpay_order_id}
                    </td>
                    <td className="px-4 py-3 text-[var(--color-text-muted)]">
                      {new Date(p.created_at).toLocaleDateString()}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Credit transactions table */}
      {tab === 'credits' && (
        <div className="mt-4 overflow-x-auto rounded-lg border border-edge">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-edge bg-overlay text-left">
                <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">User ID</th>
                <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Type</th>
                <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Amount</th>
                <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Balance After</th>
                <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Reason</th>
                <th className="px-4 py-3 font-medium text-[var(--color-text-muted)]">Date</th>
              </tr>
            </thead>
            <tbody>
              {txnsLoading ? (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-[var(--color-text-muted)]">
                    <div className="flex items-center justify-center">
                      <div className="h-5 w-5 animate-spin rounded-full border-2 border-[var(--color-primary)] border-t-transparent" />
                    </div>
                  </td>
                </tr>
              ) : txns.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-[var(--color-text-muted)]">
                    No credit transactions found.
                  </td>
                </tr>
              ) : (
                txns.map((t) => (
                  <tr key={t.id} className="border-b border-edge hover:bg-card/50">
                    <td className="px-4 py-3 font-mono text-xs text-[var(--color-text-muted)]">
                      {t.user_id.slice(0, 8)}...
                    </td>
                    <td className="px-4 py-3 text-[var(--color-text)]">{t.credit_type}</td>
                    <td className="px-4 py-3">
                      <span
                        className={
                          t.amount > 0 ? 'text-green-400' : 'text-red-400'
                        }
                      >
                        {t.amount > 0 ? '+' : ''}
                        {t.amount}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-[var(--color-text)]">{t.balance_after}</td>
                    <td className="max-w-[250px] truncate px-4 py-3 text-[var(--color-text-muted)]">
                      {t.reason}
                    </td>
                    <td className="px-4 py-3 text-[var(--color-text-muted)]">
                      {new Date(t.created_at).toLocaleDateString()}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
