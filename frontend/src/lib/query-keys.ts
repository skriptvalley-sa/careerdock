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
  lists: {
    all: ['lists'] as const,
    list: () => ['lists', 'list'] as const,
    detail: (id: string) => ['lists', 'detail', id] as const,
    entries: (listId: string) => ['lists', 'entries', listId] as const,
  },
  resumes: {
    all: ['resumes'] as const,
    list: () => ['resumes', 'list'] as const,
    detail: (id: string) => ['resumes', 'detail', id] as const,
  },
  ats: {
    all: ['ats'] as const,
    list: () => ['ats', 'list'] as const,
    detail: (id: string) => ['ats', 'detail', id] as const,
  },
  credits: {
    balance: ['credits', 'balance'] as const,
  },
  notifications: {
    all: ['notifications'] as const,
    list: () => ['notifications', 'list'] as const,
    unreadCount: () => ['notifications', 'unread-count'] as const,
  },
} as const;

// Stale time constants (milliseconds)
export const staleTimes = {
  companyList: 5 * 60 * 1000,
  companyDetail: 10 * 60 * 1000,
  userLists: 30 * 1000,
  resumes: 60 * 1000,
  atsResults: Infinity, // Immutable
  credits: 30 * 1000,
  notifications: 0, // Always fresh
} as const;
