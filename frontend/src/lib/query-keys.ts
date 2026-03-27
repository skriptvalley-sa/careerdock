// TanStack Query key factories for cache management.
// Hierarchical keys allow targeted invalidation.

export const queryKeys = {
  auth: {
    me: ['auth', 'me'] as const,
  },
  companies: {
    all: ['companies'] as const,
    list: (params?: Record<string, string>) =>
      ['companies', 'list', params] as const,
    detail: (slug: string) => ['companies', 'detail', slug] as const,
  },
  dashboard: {
    all: (userId: string) => ['dashboard', userId] as const,
    counts: (userId: string) => ['dashboard', userId, 'counts'] as const,
  },
  lists: {
    all: (userId: string) => ['lists', userId] as const,
    list: (userId: string) => ['lists', userId, 'list'] as const,
    detail: (userId: string, id: string) => ['lists', userId, 'detail', id] as const,
    entries: (userId: string, listId: string) =>
      ['lists', userId, 'entries', listId] as const,
    byCompanyAll: (userId: string) => ['lists', userId, 'by-company'] as const,
    byCompany: (userId: string, companyId: string) =>
      ['lists', userId, 'by-company', companyId] as const,
    companyCounts: (userId: string) => ['lists', userId, 'company-counts'] as const,
  },
  resumes: {
    all: (userId: string) => ['resumes', userId] as const,
    list: (userId: string) => ['resumes', userId, 'list'] as const,
    detail: (userId: string, id: string) => ['resumes', userId, 'detail', id] as const,
  },
  ats: {
    all: (userId: string) => ['ats', userId] as const,
    list: (userId: string) => ['ats', userId, 'list'] as const,
    detail: (userId: string, id: string) => ['ats', userId, 'detail', id] as const,
  },
  curatedLists: {
    all: (userId: string) => ['curated-lists', userId] as const,
    list: (userId: string) => ['curated-lists', userId, 'list'] as const,
    detail: (userId: string, id: string) =>
      ['curated-lists', userId, 'detail', id] as const,
  },
  credits: {
    all: (userId: string) => ['credits', userId] as const,
    balance: (userId: string) => ['credits', userId, 'balance'] as const,
    transactions: (userId: string) => ['credits', userId, 'transactions'] as const,
  },
  payments: {
    all: (userId: string) => ['payments', userId] as const,
    list: (userId: string) => ['payments', userId, 'list'] as const,
  },
  notifications: {
    all: (userId: string) => ['notifications', userId] as const,
    list: (userId: string) => ['notifications', userId, 'list'] as const,
    unreadCount: (userId: string) => ['notifications', userId, 'unread-count'] as const,
  },
  applications: {
    all: (userId: string) => ['applications', userId] as const,
    list: (userId: string, status?: string) =>
      ['applications', userId, 'list', status] as const,
    byCompany: (userId: string, companyId: string) =>
      ['applications', userId, 'by-company', companyId] as const,
    detail: (userId: string, id: string) =>
      ['applications', userId, 'detail', id] as const,
  },
  companyEntries: {
    all: (userId: string) => ['company-entries', userId] as const,
    byCompany: (userId: string, companyId: string) =>
      ['company-entries', userId, 'by-company', companyId] as const,
  },
  moderator: {
    all: ['moderator'] as const,
    editLock: (companyId: string) => ['moderator', 'edit-lock', companyId] as const,
  },
  admin: {
    all: ['admin'] as const,
    users: (params?: Record<string, string>) =>
      ['admin', 'users', params] as const,
    payments: (params?: Record<string, string>) =>
      ['admin', 'payments', params] as const,
    creditTransactions: (params?: Record<string, string>) =>
      ['admin', 'credit-transactions', params] as const,
  },
} as const;

// Stale time constants (milliseconds)
export const staleTimes = {
  companyList: 5 * 60 * 1000,
  companyDetail: 10 * 60 * 1000,
  userLists: 30 * 1000,
  resumes: 60 * 1000,
  atsResults: Infinity, // Immutable once complete
  curatedLists: Infinity, // Immutable once complete
  credits: 0, // always refetch on mount so balance is never stale after admin grants
  notifications: 0, // Always fresh
} as const;
