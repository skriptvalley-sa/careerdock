// CareerDock Service Worker — offline caching for the company directory.
// Served from /public/sw.js so it has root scope.

const CACHE_NAME = 'careerdock-v1';
const DB_NAME = 'careerdock';
const DB_VERSION = 1;
const COMPANY_STORE = 'companies';

// ---- IndexedDB helpers ----

function openDB() {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION);
    req.onupgradeneeded = () => {
      const db = req.result;
      if (!db.objectStoreNames.contains(COMPANY_STORE)) {
        const store = db.createObjectStore(COMPANY_STORE, { keyPath: 'slug' });
        store.createIndex('name', 'name', { unique: false });
        store.createIndex('updated_at', 'updated_at', { unique: false });
      }
    };
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

async function putCompanies(companies) {
  const db = await openDB();
  const tx = db.transaction(COMPANY_STORE, 'readwrite');
  const store = tx.objectStore(COMPANY_STORE);
  for (const c of companies) {
    store.put(c);
  }
  return new Promise((resolve, reject) => {
    tx.oncomplete = () => { db.close(); resolve(); };
    tx.onerror = () => { db.close(); reject(tx.error); };
  });
}

async function getAllCompanies() {
  const db = await openDB();
  const tx = db.transaction(COMPANY_STORE, 'readonly');
  const store = tx.objectStore(COMPANY_STORE);
  return new Promise((resolve, reject) => {
    const req = store.getAll();
    req.onsuccess = () => { db.close(); resolve(req.result); };
    req.onerror = () => { db.close(); reject(req.error); };
  });
}

async function getCompanyBySlug(slug) {
  const db = await openDB();
  const tx = db.transaction(COMPANY_STORE, 'readonly');
  const store = tx.objectStore(COMPANY_STORE);
  return new Promise((resolve, reject) => {
    const req = store.get(slug);
    req.onsuccess = () => { db.close(); resolve(req.result); };
    req.onerror = () => { db.close(); reject(req.error); };
  });
}

// ---- Service Worker lifecycle ----

self.addEventListener('install', (event) => {
  // Activate immediately — don't wait for old SW to stop
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((names) =>
      Promise.all(
        names
          .filter((name) => name !== CACHE_NAME)
          .map((name) => caches.delete(name)),
      ),
    ).then(() => self.clients.claim()),
  );
});

// ---- Fetch interception ----

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);

  // Only intercept company API calls
  if (url.pathname === '/api/companies') {
    event.respondWith(handleCompanyList(event.request));
    return;
  }

  const slugMatch = url.pathname.match(/^\/api\/companies\/([a-z0-9-]+)$/);
  if (slugMatch) {
    event.respondWith(handleCompanyDetail(event.request, slugMatch[1]));
    return;
  }
});

async function handleCompanyList(request) {
  try {
    const response = await fetch(request);
    if (response.ok) {
      const clone = response.clone();
      // Store companies in IndexedDB in the background
      clone.json().then((body) => {
        if (body.data && Array.isArray(body.data)) {
          putCompanies(body.data).catch(() => {});
        }
      });
      return response;
    }
    throw new Error('Network response not ok');
  } catch {
    // Offline fallback — serve from IndexedDB
    const companies = await getAllCompanies();
    if (companies.length === 0) {
      return new Response(
        JSON.stringify({ error: { code: 'OFFLINE', message: 'No cached data available' } }),
        { status: 503, headers: { 'Content-Type': 'application/json' } },
      );
    }
    return new Response(
      JSON.stringify({
        data: companies,
        pagination: { next_cursor: '', has_more: false },
      }),
      { status: 200, headers: { 'Content-Type': 'application/json' } },
    );
  }
}

async function handleCompanyDetail(request, slug) {
  try {
    const response = await fetch(request);
    if (response.ok) {
      const clone = response.clone();
      clone.json().then((body) => {
        if (body.data) {
          putCompanies([body.data]).catch(() => {});
        }
      });
      return response;
    }
    throw new Error('Network response not ok');
  } catch {
    // Offline fallback
    const company = await getCompanyBySlug(slug);
    if (!company) {
      return new Response(
        JSON.stringify({ error: { code: 'OFFLINE', message: 'Company not cached' } }),
        { status: 503, headers: { 'Content-Type': 'application/json' } },
      );
    }
    return new Response(
      JSON.stringify({ data: company }),
      { status: 200, headers: { 'Content-Type': 'application/json' } },
    );
  }
}
