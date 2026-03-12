# CareerDock — Frontend Design (LLD)

> **Version:** 1.0
> **Status:** Draft (Phase 3)
> **Last updated:** 2026-03-12
> **Depends on:** [PRD.md](../PRD.md), [ARCHITECTURE.md](../ARCHITECTURE.md), [api.md](./api.md)

---

## 1. Overview

The CareerDock frontend is a **Next.js 14+** application using the App Router. Public company pages are server-side rendered for SEO. Authenticated pages are client-side rendered with data fetching via React hooks. Offline browsing of previously viewed companies is supported via Service Worker + IndexedDB.

**Key decisions:**

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Framework | Next.js (App Router) | SSR for SEO, file-based routing, React Server Components |
| Styling | Tailwind CSS + shadcn/ui | Utility-first, consistent design system, accessible components |
| State management | Zustand | Lightweight, no boilerplate, works well with Next.js |
| Data fetching | TanStack Query (React Query) | Caching, background refetch, optimistic updates, SSE integration |
| Forms | React Hook Form + Zod | Performant, schema-based validation matching backend |
| HTTP client | Fetch API (wrapper) | Native, no extra dependency, cookie auth works automatically |
| Offline | Service Worker + IndexedDB (via idb) | Company directory offline browsing |

---

## 2. Route Structure

### 2.1 Public Routes (No Auth)

| Route | Page | Rendering | Description |
|-------|------|-----------|-------------|
| `/` | Landing | SSR | Value proposition, CTA to directory |
| `/companies` | Directory | SSR | Search, filter, sort company list |
| `/companies/[slug]` | Company Profile | SSR | Full company detail (SEO-optimized) |
| `/pricing` | Pricing | SSR | Starter Pack + à la carte pricing |
| `/login` | Login | CSR | Email + password login |
| `/register` | Register | CSR | Registration form |
| `/forgot-password` | Forgot Password | CSR | Request password reset |
| `/reset-password/[token]` | Reset Password | CSR | Reset form |
| `/verify-email/[token]` | Email Verify | CSR | Verification handler |

### 2.2 Authenticated Routes (Free + Premium)

| Route | Page | Auth | Description |
|-------|------|------|-------------|
| `/dashboard` | Dashboard | `authenticated` | Overview (free: lists funnel; premium: + resume health, credits) |
| `/lists` | Lists | `authenticated` | List management |
| `/lists/[id]` | List Detail | `authenticated` | Entries with application tracking |
| `/settings` | Settings | `authenticated` | Account settings |
| `/settings/profile` | Profile | `authenticated` | Edit profile/preferences |
| `/settings/payments` | Payment History | `authenticated` | Transaction history |
| `/notifications` | Notifications | `authenticated` | Notification history |

### 2.3 Premium Routes

| Route | Page | Auth | Description |
|-------|------|------|-------------|
| `/resumes` | Resume Manager | `premium` | Upload, view, set default |
| `/resumes/[id]` | Resume Detail | `premium` | Parsed data + ATS score |
| `/ats` | ATS Check | `premium` | Select check type, company/JD, resume |
| `/ats/[id]` | ATS Result | `premium` | Check result detail |
| `/curated-lists` | Curated Lists | `premium` | AI-curated list results |
| `/curated-lists/[id]` | Curated Detail | `premium` | Full ranked company list |

### 2.4 Admin Routes

| Route | Page | Auth | Description |
|-------|------|------|-------------|
| `/admin` | Admin Dashboard | `admin` | Overview stats |
| `/admin/users` | User Management | `admin` | User list, search, actions |
| `/admin/users/[id]` | User Detail | `admin` | Full user profile + credits + activity |
| `/admin/companies` | Company Management | `admin` | Company CRUD |
| `/admin/companies/new` | Create Company | `admin` | Company creation form |
| `/admin/companies/[id]/edit` | Edit Company | `admin` | Company edit form |
| `/admin/moderation` | Moderation Queue | `admin` | Pending company edits |
| `/admin/moderation/[id]` | Edit Review | `admin` | Approve/reject with diff view |
| `/admin/payments` | Payment Management | `admin` | Transaction log, refunds |
| `/admin/feature-flags` | Feature Flags | `admin` | Toggle flags |
| `/admin/ai` | AI Costs | `admin` | Cost tracking dashboard |
| `/admin/audit-log` | Audit Log | `admin` | Admin action history |

---

## 3. Directory Structure

```
frontend/
├── src/
│   ├── app/                         # Next.js App Router
│   │   ├── layout.tsx               # Root layout (providers, nav)
│   │   ├── page.tsx                 # Landing page (/)
│   │   ├── (public)/                # Public route group
│   │   │   ├── companies/
│   │   │   │   ├── page.tsx         # Directory listing
│   │   │   │   └── [slug]/
│   │   │   │       └── page.tsx     # Company profile (SSR)
│   │   │   ├── pricing/
│   │   │   │   └── page.tsx
│   │   │   ├── login/
│   │   │   │   └── page.tsx
│   │   │   ├── register/
│   │   │   │   └── page.tsx
│   │   │   ├── forgot-password/
│   │   │   │   └── page.tsx
│   │   │   ├── reset-password/
│   │   │   │   └── [token]/
│   │   │   │       └── page.tsx
│   │   │   └── verify-email/
│   │   │       └── [token]/
│   │   │           └── page.tsx
│   │   ├── (authenticated)/         # Authenticated route group
│   │   │   ├── layout.tsx           # Auth guard + sidebar layout
│   │   │   ├── dashboard/
│   │   │   │   └── page.tsx
│   │   │   ├── lists/
│   │   │   │   ├── page.tsx
│   │   │   │   └── [id]/
│   │   │   │       └── page.tsx
│   │   │   ├── notifications/
│   │   │   │   └── page.tsx
│   │   │   ├── settings/
│   │   │   │   ├── page.tsx
│   │   │   │   ├── profile/
│   │   │   │   │   └── page.tsx
│   │   │   │   └── payments/
│   │   │   │       └── page.tsx
│   │   │   ├── resumes/             # Premium-gated
│   │   │   │   ├── page.tsx
│   │   │   │   └── [id]/
│   │   │   │       └── page.tsx
│   │   │   ├── ats/                 # Premium-gated
│   │   │   │   ├── page.tsx
│   │   │   │   └── [id]/
│   │   │   │       └── page.tsx
│   │   │   └── curated-lists/       # Premium-gated
│   │   │       ├── page.tsx
│   │   │       └── [id]/
│   │   │           └── page.tsx
│   │   └── (admin)/                 # Admin route group
│   │       ├── layout.tsx           # Admin guard + admin sidebar
│   │       └── admin/
│   │           ├── page.tsx         # Dashboard
│   │           ├── users/
│   │           ├── companies/
│   │           ├── moderation/
│   │           ├── payments/
│   │           ├── feature-flags/
│   │           ├── ai/
│   │           └── audit-log/
│   ├── components/
│   │   ├── ui/                      # shadcn/ui base components
│   │   │   ├── button.tsx
│   │   │   ├── card.tsx
│   │   │   ├── dialog.tsx
│   │   │   ├── dropdown-menu.tsx
│   │   │   ├── input.tsx
│   │   │   ├── select.tsx
│   │   │   ├── table.tsx
│   │   │   ├── badge.tsx
│   │   │   ├── toast.tsx
│   │   │   ├── skeleton.tsx
│   │   │   └── ... (other shadcn primitives)
│   │   ├── layout/
│   │   │   ├── navbar.tsx           # Top navigation bar
│   │   │   ├── sidebar.tsx          # Authenticated sidebar nav
│   │   │   ├── admin-sidebar.tsx    # Admin sidebar nav
│   │   │   ├── footer.tsx
│   │   │   └── mobile-nav.tsx       # Mobile hamburger menu
│   │   ├── auth/
│   │   │   ├── login-form.tsx
│   │   │   ├── register-form.tsx
│   │   │   ├── forgot-password-form.tsx
│   │   │   └── reset-password-form.tsx
│   │   ├── companies/
│   │   │   ├── company-card.tsx     # Card in directory grid
│   │   │   ├── company-filters.tsx  # Filter sidebar/drawer
│   │   │   ├── company-search.tsx   # Search input with debounce
│   │   │   ├── company-grid.tsx     # Responsive grid layout
│   │   │   ├── company-profile.tsx  # Full profile view
│   │   │   ├── interview-patterns.tsx
│   │   │   ├── compensation-bands.tsx
│   │   │   ├── tech-stack-badges.tsx
│   │   │   └── add-to-list-button.tsx
│   │   ├── lists/
│   │   │   ├── list-card.tsx
│   │   │   ├── list-entries-table.tsx
│   │   │   ├── entry-status-badge.tsx
│   │   │   ├── entry-status-select.tsx
│   │   │   ├── add-entry-dialog.tsx
│   │   │   ├── status-history-timeline.tsx
│   │   │   └── interview-rounds-panel.tsx
│   │   ├── resumes/
│   │   │   ├── resume-upload.tsx     # Drag-and-drop upload
│   │   │   ├── resume-card.tsx       # Slot card with status
│   │   │   ├── resume-detail.tsx     # Parsed data display
│   │   │   ├── ats-score-gauge.tsx   # Circular score gauge
│   │   │   └── ats-breakdown.tsx     # Score breakdown card
│   │   ├── ats/
│   │   │   ├── ats-check-form.tsx    # Company or job check form
│   │   │   ├── ats-result-card.tsx   # Result summary card
│   │   │   ├── ats-result-detail.tsx # Full result with breakdown
│   │   │   ├── skill-match-matrix.tsx
│   │   │   └── ats-history-table.tsx
│   │   ├── curated/
│   │   │   ├── curated-list-result.tsx
│   │   │   └── company-match-card.tsx
│   │   ├── payments/
│   │   │   ├── pricing-card.tsx
│   │   │   ├── checkout-button.tsx   # Triggers Razorpay
│   │   │   ├── credit-balance.tsx    # Credit display widget
│   │   │   └── payment-history-table.tsx
│   │   ├── dashboard/
│   │   │   ├── free-dashboard.tsx
│   │   │   ├── premium-dashboard.tsx
│   │   │   ├── application-funnel.tsx
│   │   │   ├── activity-feed.tsx
│   │   │   └── resume-health-cards.tsx
│   │   ├── notifications/
│   │   │   ├── notification-bell.tsx  # Header bell with badge
│   │   │   ├── notification-list.tsx
│   │   │   └── notification-item.tsx
│   │   └── admin/
│   │       ├── stats-card.tsx
│   │       ├── user-table.tsx
│   │       ├── company-form.tsx
│   │       ├── moderation-diff.tsx   # Side-by-side diff view
│   │       ├── payment-table.tsx
│   │       ├── feature-flag-toggle.tsx
│   │       ├── ai-cost-chart.tsx
│   │       └── audit-log-table.tsx
│   ├── hooks/
│   │   ├── use-auth.ts              # Auth state + guards
│   │   ├── use-user.ts              # Current user data
│   │   ├── use-companies.ts         # Company list query
│   │   ├── use-company.ts           # Single company query
│   │   ├── use-lists.ts             # User lists queries/mutations
│   │   ├── use-resumes.ts           # Resume queries/mutations
│   │   ├── use-ats.ts               # ATS check queries/mutations
│   │   ├── use-curated-lists.ts     # Curated list queries
│   │   ├── use-credits.ts           # Credit balance + transactions
│   │   ├── use-notifications.ts     # Notification list + SSE
│   │   ├── use-sse.ts               # SSE connection hook
│   │   ├── use-debounce.ts          # Search debounce
│   │   └── use-offline.ts           # Offline detection
│   ├── lib/
│   │   ├── api.ts                   # API client (fetch wrapper)
│   │   ├── api-types.ts             # TypeScript types matching API schemas
│   │   ├── constants.ts             # App constants
│   │   ├── utils.ts                 # Utility functions
│   │   └── razorpay.ts              # Razorpay checkout helper
│   ├── store/
│   │   ├── auth-store.ts            # Auth state (Zustand)
│   │   ├── notification-store.ts    # Unread count, recent notifications
│   │   └── offline-store.ts         # Offline cached data
│   ├── workers/
│   │   └── sw.ts                    # Service worker (offline caching)
│   └── styles/
│       └── globals.css              # Tailwind imports + custom vars
├── public/
│   ├── manifest.json                # PWA manifest
│   ├── icons/                       # App icons
│   └── og/                          # Open Graph images
├── next.config.ts
├── tailwind.config.ts
├── tsconfig.json
└── package.json
```

---

## 4. State Management

### 4.1 Zustand Stores

Three small stores for global state that doesn't fit into React Query:

**Auth Store:**

```typescript
interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isPremium: boolean;
  isAdmin: boolean;
  isModerator: boolean;

  setUser: (user: User | null) => void;
  logout: () => void;
}

const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isLoading: true,
  isAuthenticated: false,
  isPremium: false,
  isAdmin: false,
  isModerator: false,

  setUser: (user) =>
    set({
      user,
      isLoading: false,
      isAuthenticated: !!user,
      isPremium: !!user?.premium_since,
      isAdmin: user?.role === 'admin',
      isModerator: user?.role === 'moderator' || user?.role === 'admin',
    }),

  logout: () =>
    set({
      user: null,
      isLoading: false,
      isAuthenticated: false,
      isPremium: false,
      isAdmin: false,
      isModerator: false,
    }),
}));
```

**Notification Store:**

```typescript
interface NotificationState {
  unreadCount: number;
  recentNotifications: Notification[];
  incrementUnread: () => void;
  setUnreadCount: (count: number) => void;
  addNotification: (notification: Notification) => void;
  markRead: (id: string) => void;
  markAllRead: () => void;
}
```

**Offline Store:**

```typescript
interface OfflineState {
  isOnline: boolean;
  cachedCompanyCount: number;
  setOnline: (online: boolean) => void;
  setCachedCount: (count: number) => void;
}
```

### 4.2 React Query (TanStack Query)

All server data fetching goes through React Query. Benefits:
- Automatic caching with configurable stale time.
- Background refetch on window focus.
- Optimistic updates for mutations.
- Loading/error states handled consistently.

**Query key conventions:**

```typescript
// Query keys follow a hierarchical pattern
const queryKeys = {
  companies: {
    all: ['companies'] as const,
    list: (filters: CompanyFilters) => ['companies', 'list', filters] as const,
    detail: (slug: string) => ['companies', 'detail', slug] as const,
  },
  lists: {
    all: ['lists'] as const,
    detail: (id: string) => ['lists', id] as const,
    entries: (listId: string) => ['lists', listId, 'entries'] as const,
  },
  resumes: {
    all: ['resumes'] as const,
    detail: (id: string) => ['resumes', id] as const,
  },
  ats: {
    all: ['ats'] as const,
    detail: (id: string) => ['ats', id] as const,
    history: (filters?: ATSFilters) => ['ats', 'history', filters] as const,
  },
  credits: {
    balance: ['credits', 'balance'] as const,
    transactions: (filters?: CreditFilters) => ['credits', 'transactions', filters] as const,
  },
  notifications: {
    all: ['notifications'] as const,
    unreadCount: ['notifications', 'unread-count'] as const,
  },
};
```

**Stale times:**

| Data | Stale Time | Rationale |
|------|-----------|-----------|
| Company list | 5 min | Changes rarely, SEO pages are SSR anyway |
| Company detail | 10 min | Same |
| User lists | 30 sec | User modifies frequently |
| Resumes | 1 min | Status changes during processing |
| ATS results | Infinity | Immutable once generated |
| Credits | 30 sec | Changes on purchase or action |
| Notifications | 0 (always fresh) | SSE handles real-time, manual refetch on page visit |

---

## 5. API Client

### 5.1 Fetch Wrapper

```typescript
// lib/api.ts

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface ApiOptions extends RequestInit {
  params?: Record<string, string | number | boolean | undefined>;
}

class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
    public details?: any,
  ) {
    super(message);
  }
}

async function api<T>(path: string, options: ApiOptions = {}): Promise<T> {
  const { params, ...fetchOptions } = options;

  // Build URL with query params
  const url = new URL(`${API_BASE}${path}`);
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined) {
        url.searchParams.set(key, String(value));
      }
    });
  }

  const response = await fetch(url.toString(), {
    ...fetchOptions,
    credentials: 'include', // Always send cookies
    headers: {
      'Content-Type': 'application/json',
      ...fetchOptions.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json();
    throw new ApiError(
      response.status,
      error.error?.code || 'UNKNOWN',
      error.error?.message || 'An error occurred',
      error.error?.details,
    );
  }

  // 204 No Content
  if (response.status === 204) {
    return undefined as T;
  }

  const json = await response.json();
  return json.data;
}

// Convenience methods
export const apiClient = {
  get: <T>(path: string, params?: Record<string, any>) =>
    api<T>(path, { method: 'GET', params }),

  post: <T>(path: string, body?: any) =>
    api<T>(path, { method: 'POST', body: JSON.stringify(body) }),

  put: <T>(path: string, body?: any) =>
    api<T>(path, { method: 'PUT', body: JSON.stringify(body) }),

  delete: <T>(path: string) =>
    api<T>(path, { method: 'DELETE' }),

  upload: <T>(path: string, formData: FormData) =>
    api<T>(path, {
      method: 'POST',
      body: formData,
      headers: {}, // Let browser set Content-Type with boundary
    }),
};
```

### 5.2 React Query Hook Example

```typescript
// hooks/use-companies.ts

export function useCompanies(filters: CompanyFilters) {
  return useInfiniteQuery({
    queryKey: queryKeys.companies.list(filters),
    queryFn: ({ pageParam }) =>
      apiClient.get<PaginatedResponse<CompanySummary>>('/api/companies', {
        ...filters,
        cursor: pageParam,
      }),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.has_more ? lastPage.pagination.next_cursor : undefined,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

export function useCompany(slug: string) {
  return useQuery({
    queryKey: queryKeys.companies.detail(slug),
    queryFn: () => apiClient.get<Company>(`/api/companies/${slug}`),
    staleTime: 10 * 60 * 1000,
  });
}
```

---

## 6. Authentication Flow (Frontend)

### 6.1 Auth Guard

```typescript
// app/(authenticated)/layout.tsx

export default function AuthenticatedLayout({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.replace('/login?redirect=' + encodeURIComponent(window.location.pathname));
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) return <LoadingSkeleton />;
  if (!isAuthenticated) return null;

  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <main className="flex-1 p-6">{children}</main>
    </div>
  );
}
```

### 6.2 Premium Guard

```typescript
// Higher-order component for premium routes
function PremiumGuard({ children }: { children: React.ReactNode }) {
  const { isPremium } = useAuthStore();

  if (!isPremium) {
    return <UpgradePrompt />;  // Shows pricing CTA
  }

  return <>{children}</>;
}
```

### 6.3 Token Refresh

```typescript
// Interceptor: auto-refresh on 401
async function apiWithRefresh<T>(path: string, options: ApiOptions = {}): Promise<T> {
  try {
    return await api<T>(path, options);
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      // Try refresh
      try {
        await api('/api/auth/refresh', { method: 'POST' });
        // Retry original request
        return await api<T>(path, options);
      } catch {
        // Refresh failed — logout
        useAuthStore.getState().logout();
        window.location.href = '/login';
        throw error;
      }
    }
    throw error;
  }
}
```

### 6.4 Initial Auth Check

On app load, check if the user has a valid session:

```typescript
// app/layout.tsx (root)
function AuthProvider({ children }: { children: React.ReactNode }) {
  const { setUser } = useAuthStore();

  useEffect(() => {
    apiClient
      .get<User>('/api/users/me')
      .then(setUser)
      .catch(() => setUser(null)); // No valid session
  }, [setUser]);

  return <>{children}</>;
}
```

---

## 7. SSE Integration

### 7.1 SSE Hook

```typescript
// hooks/use-sse.ts

export function useSSE() {
  const { isAuthenticated } = useAuthStore();
  const { addNotification, incrementUnread } = useNotificationStore();
  const queryClient = useQueryClient();

  useEffect(() => {
    if (!isAuthenticated) return;

    const eventSource = new EventSource(
      `${API_BASE}/api/notifications/stream`,
      { withCredentials: true }
    );

    eventSource.addEventListener('resume_parsed', (event) => {
      const data = JSON.parse(event.data);
      addNotification(data);
      incrementUnread();
      // Invalidate resume queries to refetch with new status
      queryClient.invalidateQueries({ queryKey: queryKeys.resumes.all });
    });

    eventSource.addEventListener('ats_result_ready', (event) => {
      const data = JSON.parse(event.data);
      addNotification(data);
      incrementUnread();
      queryClient.invalidateQueries({ queryKey: queryKeys.ats.all });
    });

    eventSource.addEventListener('curated_list_ready', (event) => {
      const data = JSON.parse(event.data);
      addNotification(data);
      incrementUnread();
      queryClient.invalidateQueries({ queryKey: ['curated-lists'] });
    });

    eventSource.addEventListener('credits_updated', (event) => {
      const data = JSON.parse(event.data);
      addNotification(data);
      incrementUnread();
      queryClient.invalidateQueries({ queryKey: queryKeys.credits.balance });
    });

    eventSource.onerror = () => {
      // EventSource auto-reconnects. Log for debugging.
      console.warn('SSE connection lost, reconnecting...');
    };

    return () => eventSource.close();
  }, [isAuthenticated]);
}
```

### 7.2 SSE Placement

The `useSSE()` hook is called once in the authenticated layout, so all authenticated pages receive real-time updates:

```typescript
// app/(authenticated)/layout.tsx
export default function AuthenticatedLayout({ children }) {
  useSSE(); // Single SSE connection for all authenticated pages
  // ...
}
```

---

## 8. Offline Support

### 8.1 Strategy

**What's available offline:**
- Previously viewed company profile pages (from IndexedDB cache).
- Company directory search results (cached page data).
- Static assets (JS, CSS, images via Service Worker).

**What's NOT available offline:**
- Authenticated features (lists, resumes, ATS).
- New searches or filters (requires API).
- Any write operations.

### 8.2 Service Worker

```typescript
// workers/sw.ts (compiled separately)

const CACHE_NAME = 'careerdock-v1';
const STATIC_ASSETS = ['/', '/pricing', '/manifest.json'];

// Install: cache static shell
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(STATIC_ASSETS))
  );
});

// Fetch: network-first for API, cache-first for static assets
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);

  // API requests: network-first, cache company responses
  if (url.pathname.startsWith('/api/companies')) {
    event.respondWith(networkFirstWithCache(event.request));
    return;
  }

  // Static assets: cache-first
  if (url.pathname.match(/\.(js|css|png|jpg|svg|woff2)$/)) {
    event.respondWith(cacheFirst(event.request));
    return;
  }

  // HTML pages: network-first
  event.respondWith(networkFirst(event.request));
});

async function networkFirstWithCache(request: Request): Promise<Response> {
  try {
    const response = await fetch(request);
    // Cache successful GET responses for company data
    if (response.ok && request.method === 'GET') {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, response.clone());
      // Also store in IndexedDB for structured access
      storeCompanyInIDB(request.url, await response.clone().json());
    }
    return response;
  } catch {
    // Offline: try cache
    const cached = await caches.match(request);
    if (cached) return cached;
    return new Response(JSON.stringify({ error: { code: 'OFFLINE', message: 'You are offline' } }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' },
    });
  }
}
```

### 8.3 IndexedDB for Company Data

```typescript
// Using 'idb' library for typed IndexedDB access

import { openDB, IDBPDatabase } from 'idb';

interface CareerDockDB {
  companies: {
    key: string; // slug
    value: Company;
    indexes: { 'by-updated': string };
  };
  searchResults: {
    key: string; // serialized filter params
    value: { companies: CompanySummary[]; cachedAt: number };
  };
}

const dbPromise = openDB<CareerDockDB>('careerdock', 1, {
  upgrade(db) {
    const companyStore = db.createObjectStore('companies', { keyPath: 'slug' });
    companyStore.createIndex('by-updated', 'updated_at');
    db.createObjectStore('searchResults');
  },
});

export async function getCachedCompany(slug: string): Promise<Company | undefined> {
  const db = await dbPromise;
  return db.get('companies', slug);
}

export async function cacheCompany(company: Company): Promise<void> {
  const db = await dbPromise;
  await db.put('companies', company);
}

export async function getCachedCompanyCount(): Promise<number> {
  const db = await dbPromise;
  return db.count('companies');
}
```

### 8.4 Offline Detection Hook

```typescript
// hooks/use-offline.ts

export function useOffline() {
  const { isOnline, setOnline, cachedCompanyCount, setCachedCount } = useOfflineStore();

  useEffect(() => {
    const handleOnline = () => setOnline(true);
    const handleOffline = () => setOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);
    setOnline(navigator.onLine);

    // Count cached companies
    getCachedCompanyCount().then(setCachedCount);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  return { isOnline, cachedCompanyCount };
}
```

---

## 9. Key Page Designs

### 9.1 Company Directory (`/companies`)

```
┌─────────────────────────────────────────────────────────────┐
│  CareerDock                       [Login] [Register]        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  🔍 Search companies...                    [Filters ▼]     │
│                                                             │
│  ┌─── Filters (collapsible) ───────────────────────────┐   │
│  │ Tech Stack: [Go] [Python] [Java] [React] ...        │   │
│  │ Domain: [Cloud] [FinTech] [SaaS] [AI/ML] ...        │   │
│  │ Size: [Startup] [Small] [Mid] [Large] [Enterprise]  │   │
│  │ Tier: [Tier 1] [Tier 2] [Tier 3] [Tier 4]          │   │
│  │ Hiring: [Active] [Paused] [Unknown]                  │   │
│  │ RSU: [Has RSU]                                       │   │
│  │ Sort by: [Name ▼]                    [Clear filters] │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                             │
│  Showing 156 companies                                      │
│                                                             │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐        │
│  │ 🏢 Google     │ │ 🏢 Microsoft  │ │ 🏢 Flipkart  │        │
│  │ Enterprise   │ │ Enterprise   │ │ Large        │        │
│  │ Bangalore    │ │ Hyderabad    │ │ Bangalore    │        │
│  │              │ │              │ │              │        │
│  │ Go C++ K8s   │ │ C# .NET K8s │ │ Java Go React│        │
│  │              │ │              │ │              │        │
│  │ Tier 1  🟢   │ │ Tier 1  🟢   │ │ Tier 4  🟢   │        │
│  │ RSU ✓        │ │ RSU ✓        │ │ ESOP         │        │
│  │ [+ Add]      │ │ [+ Add]      │ │ [+ Add]      │        │
│  └──────────────┘ └──────────────┘ └──────────────┘        │
│                                                             │
│  [Load more...]                                             │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

- Responsive: 3-column grid on desktop, 2 on tablet, 1 on mobile.
- Infinite scroll with cursor pagination.
- Filter chips are toggleable. URL params update on filter change (shareable URLs).
- "Add" button opens a list selector (for authenticated users).

### 9.2 Company Profile (`/companies/[slug]`)

```
┌─────────────────────────────────────────────────────────────┐
│  ← Back to Directory                                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  🏢 Google                                     Tier 1      │
│  Enterprise · Bangalore, Karnataka · Founded 1998           │
│  🟢 Actively Hiring                                         │
│                                                             │
│  [Careers Page] [Glassdoor] [AmbitionBox] [LinkedIn]        │
│  [+ Add to List]                                            │
│                                                             │
│  ──── About ────────────────────────────────────────        │
│  Global technology company specializing in...                │
│                                                             │
│  ──── Tech Stack ───────────────────────────────────        │
│  [Go] [C++] [Python] [gRPC] [Kubernetes] [Spanner]         │
│                                                             │
│  ──── Domains ──────────────────────────────────────        │
│  [Cloud] [AI/ML] [Infra] [Platform]                         │
│                                                             │
│  ──── Interview Process ────────────────────────────        │
│  ┌─ SDE-1 (4 rounds, ~14 days) ─────────────────┐          │
│  │ 1. Online Assessment (90 min) - DSA, Medium   │          │
│  │ 2. Technical Interview (60 min) - DSA, M-Hard │          │
│  │ 3. System Design (45 min) - Basics            │          │
│  │ 4. Hiring Manager (30 min) - Behavioral       │          │
│  └───────────────────────────────────────────────┘          │
│  ┌─ SDE-2 (5 rounds, ~21 days) ...              ┘          │
│                                                             │
│  ──── Compensation ─────────────────────────────────        │
│  RSU: ✓ Yes   Refreshers: ✓ Yes                            │
│  ┌───────────┬──────────────┬──────────────────┐            │
│  │ Role      │ Range (LPA)  │ Equity           │            │
│  ├───────────┼──────────────┼──────────────────┤            │
│  │ SDE-1     │ ₹25-40L      │ RSU, 4yr vest    │            │
│  │ SDE-2     │ ₹40-60L      │ RSU, refreshers  │            │
│  │ SDE-3     │ ₹60-90L      │ RSU, refreshers  │            │
│  └───────────┴──────────────┴──────────────────┘            │
│  ⚠ Compensation data are community estimates.               │
│                                                             │
│  Last verified: March 1, 2026                               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 9.3 Dashboard (`/dashboard`)

**Free user:**

```
┌───────────────────────────────────────────────────────┐
│  Dashboard                                             │
├───────────────────────────────────────────────────────┤
│                                                       │
│  ┌─ My Lists (2/3) ──────────────────────────────┐   │
│  │ Dream Companies (8 entries)        [View →]    │   │
│  │ Backup Options (4 entries)         [View →]    │   │
│  │ [+ Create New List]                            │   │
│  └────────────────────────────────────────────────┘   │
│                                                       │
│  ┌─ Application Funnel ──────────────────────────┐   │
│  │ Not Applied  ████████████████  8               │   │
│  │ Applied      ████████         4               │   │
│  │ Interview    ████             2               │   │
│  │ Offer        ██               1               │   │
│  │ Rejected     ███              1               │   │
│  └────────────────────────────────────────────────┘   │
│                                                       │
│  ┌─ Recent Activity ─────────────────────────────┐   │
│  │ • Applied to Google (SDE-3)        2 hours ago │   │
│  │ • Added Microsoft to Dream List    1 day ago   │   │
│  │ • Created "Backup Options" list    3 days ago  │   │
│  └────────────────────────────────────────────────┘   │
│                                                       │
│  ┌─ 🚀 Upgrade to Premium ──────────────────────┐   │
│  │ Get AI-powered resume analysis, ATS scoring,  │   │
│  │ and curated company recommendations.           │   │
│  │                     [Get Started — ₹399 →]     │   │
│  └────────────────────────────────────────────────┘   │
│                                                       │
└───────────────────────────────────────────────────────┘
```

**Premium user** adds resume health cards, credit tracker, and curated list summary above the lists section.

---

## 10. SEO Strategy

### 10.1 Server-Side Rendering

Company pages use Next.js Server Components for SSR:

```typescript
// app/(public)/companies/[slug]/page.tsx

import { Metadata } from 'next';

// Generate metadata for SEO
export async function generateMetadata({ params }): Promise<Metadata> {
  const company = await fetchCompany(params.slug);
  return {
    title: `${company.name} — Tech Stack, Interviews & Compensation | CareerDock`,
    description: `${company.name} engineering info: tech stack (${company.tech_stack.slice(0, 5).join(', ')}), interview process, compensation bands, and hiring status.`,
    openGraph: {
      title: company.name,
      description: company.description,
      type: 'website',
      url: `https://careerdock.skriptvalley.com/companies/${company.slug}`,
    },
  };
}

// JSON-LD structured data
function CompanyJsonLd({ company }: { company: Company }) {
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'Organization',
    name: company.name,
    url: company.careers_page_url,
    location: {
      '@type': 'Place',
      address: company.headquarters,
    },
    foundingDate: company.founded_year?.toString(),
    description: company.description,
  };

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
    />
  );
}
```

### 10.2 Sitemap

Auto-generated sitemap including all company profile URLs:

```typescript
// app/sitemap.ts
export default async function sitemap() {
  const companies = await fetchAllCompanySlugs();
  return [
    { url: 'https://careerdock.skriptvalley.com', changeFrequency: 'weekly' },
    { url: 'https://careerdock.skriptvalley.com/companies', changeFrequency: 'daily' },
    { url: 'https://careerdock.skriptvalley.com/pricing', changeFrequency: 'monthly' },
    ...companies.map((slug) => ({
      url: `https://careerdock.skriptvalley.com/companies/${slug}`,
      changeFrequency: 'weekly' as const,
    })),
  ];
}
```

### 10.3 Cache Headers

| Route | Cache-Control |
|-------|--------------|
| `/companies` | `public, s-maxage=300, stale-while-revalidate=60` |
| `/companies/[slug]` | `public, s-maxage=600, stale-while-revalidate=120` |
| `/pricing` | `public, s-maxage=3600` |
| Authenticated pages | `private, no-cache` |
| API responses (public) | `public, max-age=300` + ETag |

---

## 11. Responsive Design

### 11.1 Breakpoints (Tailwind defaults)

| Breakpoint | Min Width | Target |
|-----------|-----------|--------|
| `sm` | 640px | Large phones (landscape) |
| `md` | 768px | Tablets |
| `lg` | 1024px | Small laptops |
| `xl` | 1280px | Desktops |

### 11.2 Mobile Adaptations

| Component | Desktop | Mobile |
|-----------|---------|--------|
| Navigation | Horizontal top bar | Hamburger menu |
| Sidebar | Fixed left sidebar | Bottom sheet / hidden |
| Company grid | 3 columns | 1 column, card layout |
| Filters | Sidebar panel | Bottom drawer |
| Tables | Full table | Card list or horizontal scroll |
| Dialogs | Center modal | Full-screen sheet |

### 11.3 Mobile-First Approach

All components are styled mobile-first with Tailwind:

```tsx
// Example: Company grid
<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
  {companies.map((company) => (
    <CompanyCard key={company.id} company={company} />
  ))}
</div>
```

---

## 12. Razorpay Checkout Integration

```typescript
// lib/razorpay.ts

declare global {
  interface Window {
    Razorpay: any;
  }
}

interface CheckoutOptions {
  orderId: string;
  amount: number;
  keyId: string;
  user: { email: string; name: string };
  onSuccess: () => void;
  onFailure: (error: any) => void;
}

export function openRazorpayCheckout(options: CheckoutOptions) {
  const rzp = new window.Razorpay({
    key: options.keyId,
    amount: options.amount,
    currency: 'INR',
    name: 'CareerDock',
    description: 'Premium Features',
    order_id: options.orderId,
    handler: () => {
      // Payment successful on client side
      // Don't allocate credits — wait for webhook via SSE
      options.onSuccess();
    },
    prefill: {
      email: options.user.email,
      name: options.user.name,
    },
    theme: { color: '#4F46E5' },
    modal: {
      ondismiss: () => {
        // User closed checkout without paying
      },
    },
  });

  rzp.on('payment.failed', (response: any) => {
    options.onFailure(response.error);
  });

  rzp.open();
}
```

**Script loading:** Razorpay checkout script loaded on demand (not in initial bundle):

```typescript
// Load Razorpay script only when needed
export function loadRazorpayScript(): Promise<void> {
  return new Promise((resolve, reject) => {
    if (window.Razorpay) {
      resolve();
      return;
    }
    const script = document.createElement('script');
    script.src = 'https://checkout.razorpay.com/v1/checkout.js';
    script.onload = () => resolve();
    script.onerror = () => reject(new Error('Failed to load Razorpay'));
    document.head.appendChild(script);
  });
}
```

---

## 13. Error Handling

### 13.1 Global Error Boundary

```typescript
// app/error.tsx
'use client';

export default function GlobalError({ error, reset }: { error: Error; reset: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[50vh] gap-4">
      <h2 className="text-xl font-semibold">Something went wrong</h2>
      <p className="text-muted-foreground">{error.message}</p>
      <Button onClick={reset}>Try again</Button>
    </div>
  );
}
```

### 13.2 API Error Handling in Hooks

```typescript
// Centralized error handling via React Query's onError
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => {
        if (error instanceof ApiError) {
          // Don't retry on auth or validation errors
          if ([401, 403, 404, 422].includes(error.status)) return false;
          // Retry server errors up to 2 times
          if (error.status >= 500) return failureCount < 2;
        }
        return failureCount < 2;
      },
    },
    mutations: {
      onError: (error) => {
        if (error instanceof ApiError) {
          // Show toast for user-facing errors
          toast.error(error.message);
        }
      },
    },
  },
});
```

### 13.3 Form Validation Errors

Using Zod schemas that mirror backend validation:

```typescript
import { z } from 'zod';

export const registerSchema = z.object({
  email: z.string().email('Invalid email address').max(255),
  password: z
    .string()
    .min(8, 'Password must be at least 8 characters')
    .regex(/[A-Z]/, 'Must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Must contain at least one number'),
  name: z.string().min(1, 'Name is required').max(255),
});

export const companyFiltersSchema = z.object({
  q: z.string().optional(),
  tech_stack: z.string().optional(),
  domains: z.string().optional(),
  size: z.string().optional(),
  hiring_status: z.enum(['active', 'paused', 'unknown']).optional(),
  compensation_tier: z.string().optional(),
  has_rsu: z.boolean().optional(),
});
```

---

## 14. Performance

### 14.1 Bundle Optimization

- **Route-based code splitting** — automatic with Next.js App Router.
- **Dynamic imports** for heavy components:

```typescript
const InterviewPatterns = dynamic(() => import('@/components/companies/interview-patterns'), {
  loading: () => <Skeleton className="h-48" />,
});

const ATSBreakdown = dynamic(() => import('@/components/ats/ats-breakdown'));
```

- **Razorpay script** loaded on demand (see §12).
- **Admin pages** in separate route group — never loaded for regular users.

### 14.2 Image Optimization

- Company logos: served via Next.js `<Image>` with automatic resizing + WebP conversion.
- Fallback: generated initials avatar for companies without logos.

```typescript
function CompanyLogo({ company }: { company: CompanySummary }) {
  if (company.logo_url) {
    return (
      <Image
        src={company.logo_url}
        alt={company.name}
        width={48}
        height={48}
        className="rounded"
      />
    );
  }
  return (
    <div className="w-12 h-12 rounded bg-primary/10 flex items-center justify-center text-lg font-semibold">
      {company.name[0]}
    </div>
  );
}
```

### 14.3 Target Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| LCP (company directory) | <2.0s | Lighthouse |
| FCP (landing page) | <1.5s | Lighthouse |
| TTI (authenticated pages) | <3.5s | Lighthouse |
| CLS | <0.1 | Lighthouse |
| JS bundle (initial) | <150 KB gzipped | Build output |

---

## 15. Cross-Reference

| Architecture Decision | Frontend Implementation |
|----------------------|----------------------|
| Next.js App Router (§3.1) | SSR for public pages, CSR for authenticated |
| JWT in httpOnly cookies (§5) | `credentials: 'include'` on all fetch calls, no token handling in JS |
| SSE for notifications (§3.10) | `useSSE()` hook with EventSource, auto-reconnection |
| Offline support (§3.1.3) | Service Worker + IndexedDB via `idb` library |
| Razorpay integration (§3.8) | Dynamic script loading, checkout widget wrapper |
| Tailwind + shadcn/ui (§3.1) | Utility-first CSS, accessible component library |
| Client-side caching (§3.3) | React Query stale times + Service Worker + IndexedDB |
