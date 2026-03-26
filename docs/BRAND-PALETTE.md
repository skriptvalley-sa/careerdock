# Brand Palette — Implementation Guide

> **Session goal:** Replace the current neon cyberpunk dark-only theme with a dual-mode
> palette (light + dark) that matches the approved brand reference images.
>
> Reference images: `docs/design-references/color-pallets/combo_05.jpeg` (light)
> and `docs/design-references/color-pallets/combo_03.jpeg` (dark)

---

## Approved Palette

| Mode | Role | Name | Hex |
|------|------|------|-----|
| Light | Background / surface | Lemon Chiffon | `#FEFACD` |
| Light | Primary / accent | Ultra Violet | `#5F4A8B` |
| Dark | Background / surface | Charcoal | `#233D4C` |
| Dark | Primary / accent | Pumpkin | `#FD802E` |

---

## Tech Context (read before starting)

- **Tailwind v4** — no `tailwind.config.ts`. All tokens live in `frontend/src/styles/globals.css`
  inside a `@theme { }` block. Dark-mode values go in a `.dark { }` CSS block that
  overrides the same variable names.
- **No animation library** — all animations are raw CSS keyframes in `globals.css`.
  Do **not** install `framer-motion` or `tailwindcss-animate` unless the task explicitly
  requires it; CSS keyframes are sufficient and keep the bundle lean.
- **`next-themes`** is not yet installed — it must be added to wire up the `dark` class on
  `<html>` in a hydration-safe way.
- **Existing neon utilities** (`.glow-cyan`, `.btn-neon`, `.card-neon-hover`, etc.) must be
  renamed to palette-aware equivalents and all call-sites updated.

---

## Implementation Steps

### Step 1 — Install `next-themes`

```bash
cd frontend && npm install next-themes
```

---

### Step 2 — Wire `ThemeProvider` into the app

**File: `frontend/src/app/layout.tsx`**

Add `suppressHydrationWarning` to the `<html>` tag to prevent the hydration mismatch
that occurs when `next-themes` sets the `dark` class server-side vs. client:

```tsx
<html lang="en" suppressHydrationWarning>
```

**File: `frontend/src/components/providers.tsx`**

Import and wrap everything with `ThemeProvider`. `defaultTheme="dark"` matches the
current shipped experience; users can toggle to light. `attribute="class"` makes
`next-themes` add/remove the `dark` class on `<html>`, which is what Tailwind's
`.dark` CSS block hooks into.

```tsx
import { ThemeProvider } from 'next-themes'

// Inside the Providers component, outermost wrapper:
<ThemeProvider attribute="class" defaultTheme="dark" disableTransitionOnChange={false}>
  {/* existing QueryClientProvider, AuthProvider, SidebarContext */}
</ThemeProvider>
```

---

### Step 3 — Replace `globals.css` tokens and utilities

Replace the entire contents of `frontend/src/styles/globals.css` with the following.
The `@theme` block sets light-mode defaults; the `.dark` block overrides them.

```css
@import 'tailwindcss';

/* ═══════════════════════════════════════════════════════
   BRAND PALETTE
   Light (Combo 05):  Lemon Chiffon #FEFACD + Ultra Violet #5F4A8B
   Dark  (Combo 03):  Charcoal      #233D4C + Pumpkin      #FD802E
   ═══════════════════════════════════════════════════════ */

/* ── Light-mode tokens (default) ── */
@theme {
  --color-surface:        #FEFACD;
  --color-card:           #FDF7B8;
  --color-overlay:        #FAF5A0;
  --color-input:          #FFF9D6;
  --color-primary:        #5F4A8B;
  --color-primary-hover:  #4E3A75;
  --color-text:           #2A1F4A;
  --color-text-muted:     #7A6A9E;
  --color-edge:           rgba(95, 74, 139, 0.15);
  --color-edge-input:     rgba(95, 74, 139, 0.25);
  --color-edge-hover:     rgba(95, 74, 139, 0.40);
  --color-accent:         #5F4A8B;
}

/* ── Dark-mode token overrides ── */
.dark {
  --color-surface:        #233D4C;
  --color-card:           #1C3040;
  --color-overlay:        #162837;
  --color-input:          #1E3448;
  --color-primary:        #FD802E;
  --color-primary-hover:  #E5711F;
  --color-text:           #F0E8D0;
  --color-text-muted:     #8FAABF;
  --color-edge:           rgba(253, 128, 46, 0.15);
  --color-edge-input:     rgba(253, 128, 46, 0.25);
  --color-edge-hover:     rgba(253, 128, 46, 0.40);
  --color-accent:         #FD802E;
}

/* ── Smooth mode transition ── */
body {
  background-color: var(--color-surface);
  color: var(--color-text);
  transition:
    background-color 300ms ease,
    color 300ms ease,
    border-color 300ms ease;
}

/* ── Custom scrollbar ── */
::-webkit-scrollbar { width: 6px; }
::-webkit-scrollbar-track { background: var(--color-overlay); }
::-webkit-scrollbar-thumb {
  background: var(--color-edge-hover);
  border-radius: 3px;
}
::-webkit-scrollbar-thumb:hover { background: var(--color-primary); opacity: 0.6; }

/* ── Focus ring ── */
input:focus, select:focus, textarea:focus { outline: none; }

/* ══════════════════════════════════
   UTILITY CLASSES
   ══════════════════════════════════ */

/* Primary glow (replaces .glow-cyan) */
.glow-primary {
  box-shadow:
    0 0 12px color-mix(in srgb, var(--color-primary) 30%, transparent),
    0 0  4px color-mix(in srgb, var(--color-primary) 15%, transparent);
}

/* Primary button (replaces .btn-neon) */
.btn-primary {
  background-color: var(--color-primary);
  color: var(--color-surface);
  transition: background-color 200ms ease, box-shadow 200ms ease, transform 100ms ease;
}
.btn-primary:hover {
  background-color: var(--color-primary-hover);
  box-shadow: 0 0 20px color-mix(in srgb, var(--color-primary) 40%, transparent);
}
.btn-primary:active { transform: scale(0.97); }

/* Card hover (replaces .card-neon-hover) */
.card-hover {
  transition: border-color 200ms ease, box-shadow 200ms ease;
}
.card-hover:hover {
  border-color: var(--color-edge-hover);
  box-shadow:
    0 0 16px color-mix(in srgb, var(--color-primary) 10%, transparent),
    inset 0 0 16px color-mix(in srgb, var(--color-primary) 4%, transparent);
}

/* Link hover (replaces .neon-link) */
a.primary-link:hover {
  color: var(--color-primary);
}

/* ══════════════════════════════════
   KEYFRAME ANIMATIONS
   ══════════════════════════════════ */

/* 1. CTA glow pulse — apply with class .animate-pulse-primary */
@keyframes pulse-primary {
  0%, 100% {
    box-shadow: 0 0  8px color-mix(in srgb, var(--color-primary) 25%, transparent);
  }
  50% {
    box-shadow: 0 0 20px color-mix(in srgb, var(--color-primary) 50%, transparent),
                0 0  6px color-mix(in srgb, var(--color-primary) 20%, transparent);
  }
}
.animate-pulse-primary {
  animation: pulse-primary 2.4s ease-in-out infinite;
}

/* 2. Card / section entrance — apply with class .animate-fade-up */
@keyframes fade-up {
  from { opacity: 0; transform: translateY(12px); }
  to   { opacity: 1; transform: translateY(0); }
}
.animate-fade-up {
  animation: fade-up 400ms ease forwards;
}

/* Stagger helpers */
.animate-fade-up.delay-100 { animation-delay: 100ms; }
.animate-fade-up.delay-200 { animation-delay: 200ms; }
.animate-fade-up.delay-300 { animation-delay: 300ms; }

/* 3. Sidebar / nav link accent slide */
@keyframes slide-accent {
  from { border-left-color: transparent; }
  to   { border-left-color: var(--color-primary); }
}
.nav-link-active {
  border-left: 2px solid var(--color-primary);
  animation: slide-accent 200ms ease forwards;
  color: var(--color-primary);
}

/* 4. Input focus ring */
.input-primary:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-primary) 20%, transparent);
  transition: border-color 150ms ease, box-shadow 150ms ease;
}
```

---

### Step 4 — Audit and replace hardcoded neon color classes

Search the entire `frontend/src/` directory for any Tailwind classes that reference
the old neon palette and replace them with the new semantic equivalents.

**Search patterns to find and fix:**

```bash
# Run from frontend/
grep -rn "text-cyan\|border-cyan\|bg-cyan\|ring-cyan\|text-indigo\|bg-indigo\|border-indigo\|text-purple\|bg-purple\|border-purple\|glow-cyan\|glow-magenta\|glow-green\|glow-amber\|btn-neon\|card-neon-hover\|neon-link\|text-glow" src/
```

**Replacement mapping:**

| Old class | New class / approach |
|-----------|----------------------|
| `.glow-cyan` | `.glow-primary` |
| `.btn-neon` | `.btn-primary` |
| `.card-neon-hover` | `.card-hover` |
| `.neon-link` | `.primary-link` |
| `text-cyan-400`, `text-cyan-300` | `text-[var(--color-primary)]` or `text-[var(--color-text)]` |
| `border-cyan-*` | `border-[var(--color-edge)]` or `border-[var(--color-edge-hover)]` |
| `bg-indigo-*`, `bg-purple-*` | `bg-[var(--color-primary)]` |
| `text-indigo-*`, `text-purple-*` | `text-[var(--color-primary)]` or `text-[var(--color-text-muted)]` |
| `text-glow-cyan`, `text-glow-magenta` | Remove or replace with inline style if needed |

---

### Step 5 — Create the theme toggle component

**New file: `frontend/src/components/ui/theme-toggle.tsx`**

```tsx
'use client'

import { useTheme } from 'next-themes'
import { Sun, Moon } from 'lucide-react'
import { useEffect, useState } from 'react'

export function ThemeToggle() {
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = useState(false)

  // Avoid hydration mismatch — only render icon after mount
  useEffect(() => setMounted(true), [])
  if (!mounted) return <div className="w-8 h-8" />

  return (
    <button
      onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
      className="p-2 rounded-lg border border-[var(--color-edge)] hover:border-[var(--color-edge-hover)]
                 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]
                 transition-colors duration-200"
      aria-label="Toggle theme"
    >
      {theme === 'dark' ? <Sun size={16} /> : <Moon size={16} />}
    </button>
  )
}
```

**Add the toggle to the sidebar header or top app bar** — locate the header component
(likely `frontend/src/components/layout/sidebar.tsx` or similar) and import + render
`<ThemeToggle />` in the top-right corner alongside the notification bell.

---

### Step 6 — Contrast ratio verification

Before finishing, verify both palette combinations meet WCAG AA (≥ 4.5:1 for normal text):

| Foreground | Background | Expected ratio | Pass? |
|-----------|-----------|---------------|-------|
| Ultra Violet `#5F4A8B` | Lemon Chiffon `#FEFACD` | ~5.2:1 | ✅ |
| Pumpkin `#FD802E` | Charcoal `#233D4C` | ~3.8:1 ⚠️ | Check — may need `#FF9040` for body text |

> **Note:** Pumpkin on Charcoal is borderline for small body text. Use Pumpkin only for
> large headings, icons, and interactive elements in dark mode. For body text on Charcoal,
> use `--color-text` (`#F0E8D0`) which has excellent contrast (~12:1).

---

### Step 7 — Pages to verify visually (use browser MCP)

After implementation, screenshot each page in both light and dark mode:

1. `/pricing` — CTA buttons should pulse with `.animate-pulse-primary`
2. `/dashboard` — stat cards should show `.animate-fade-up` on first load
3. `/companies` — company cards should show `.card-hover` glow on hover
4. `/admin/companies` — table + edit modal in both modes
5. `/login`, `/register`, `/forgot-password` — input focus rings
6. Toggle between modes — 300ms smooth crossfade on body background

---

## Branch & PR

```bash
git checkout -b feat/brand-palette-v2
# ... make changes, commit incrementally ...
make lint          # must pass before push
make build         # must pass before push
git push origin feat/brand-palette-v2
gh pr create --title "feat: implement Combo 05/03 brand palette with light/dark mode"
```

---

## File Change Summary

| File | Type | What changes |
|------|------|-------------|
| `frontend/package.json` | Edit | Add `next-themes` dependency |
| `frontend/src/app/layout.tsx` | Edit | Add `suppressHydrationWarning` to `<html>` |
| `frontend/src/components/providers.tsx` | Edit | Wrap with `<ThemeProvider>` |
| `frontend/src/styles/globals.css` | **Replace** | New dual-mode tokens + utilities + keyframes |
| `frontend/src/components/ui/theme-toggle.tsx` | **Create** | Sun/moon toggle button |
| `frontend/src/components/layout/sidebar.tsx` | Edit | Add `<ThemeToggle />` |
| All components with hardcoded neon classes | Edit | Replace per mapping table in Step 4 |
