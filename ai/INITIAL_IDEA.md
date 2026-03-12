# Claude Code Prompt: Bangalore Tech Companies Reference App

Build a React single-file component (`bangalore_companies.jsx`) that serves as an interactive company reference tool for software engineers job-hunting in Bangalore.

---

## Data Model

Each company entry is an object with these fields:

```ts
{
  name: string,         // Company name
  tier: 1 | 2 | 3 | 4, // Compensation tier
  domain: string,       // Short domain description e.g. "Cloud / Enterprise"
  rsu: boolean,         // Whether they grant RSUs
  refresher: boolean,   // Whether annual RSU refreshers are standard
  notes: string,        // 1–2 line comp/context note with ₹ range if known
  tags: string[],       // Domain tags e.g. ["Cloud", "Infra", "AI/ML"]
}
```

Populate with ~76 companies across 4 tiers:

- **Tier 1** — FAANG-level: Google, Microsoft, Amazon, Apple, Meta
- **Tier 2** — Top MNCs with confirmed RSU + refreshers: Adobe, Salesforce, Cisco, Visa, Broadcom, SAP Labs, Intuit, Atlassian, LinkedIn, Uber, Walmart Global Tech, PayPal, Mastercard, Amex, ServiceNow, Workday, Palo Alto Networks, CrowdStrike, Zscaler, Oracle, Intel, Qualcomm
- **Tier 3** — Solid MNCs / good equity: MongoDB, Elastic, Confluent, Databricks, Cloudflare, Okta, Twilio, Nutanix, Pure Storage, NetApp, Rubrik, Cohesity, JFrog, Dynatrace, New Relic, Marvell, Arista, Juniper, Synopsys, Cadence, Texas Instruments, ARM, Western Digital, Seagate, Micron, HPE, Dell, Honeywell, Bosch, Samsung R&D, Siemens EDA
- **Tier 4** — Indian product companies with ESOP/RSU: Flipkart, PhonePe, Swiggy, Zomato, Razorpay, Freshworks, CRED, Meesho, Groww, Zepto, InMobi, Browserstack, Clevertap

Available tags across all companies: `AI/ML`, `Cloud`, `Infra`, `Distributed`, `Platform`, `SaaS`, `FinTech`, `Security`, `Networking`, `Dev Tools`, `Embedded`, `Database`, `Storage`, `Automotive`

---

## Filtering Logic

Maintain the following filter state:

| State | Type | Default | Behaviour |
|---|---|---|---|
| `search` | `string` | `""` | Case-insensitive match against `name`, `domain`, `notes` |
| `selectedTiers` | `number[]` | `[1,2,3,4]` | Toggle individual tiers on/off |
| `selectedTags` | `string[]` | `[]` | Multi-select; show companies matching **any** selected tag |
| `rsuOnly` | `boolean` | `false` | Hide companies where `rsu === false` |
| `refresherOnly` | `boolean` | `false` | Hide companies where `refresher === false` |

Filtering is computed via `useMemo` for performance.

---

## Layout

### Sticky Header
Dark background (`#0d0f18`), `position: sticky`, `top: 0`, `z-index: 10`. Contains:

1. Title + live count label (`X of 76 companies`)
2. Search input — monospace font, dark-styled, full text search
3. Checkboxes: `RSU only` and `With Refreshers`
4. **Tier filter chips** — one per tier, shows filtered count. Click to toggle. Color-coded:
   - T1: amber · T2: green · T3: blue · T4: purple
5. **Tag filter chips** — all unique tags listed inline. Click to toggle active state. Visually distinct when active. Clicking a tag chip on a card also toggles the tag filter.

### Main Content
Grouped by tier. Only render groups that have at least one result after filtering.

Each group has:
- A section header: colored short divider line + uppercase tier label + result count
- A CSS grid of cards: `repeat(auto-fill, minmax(300px, 1fr))`

### Card Structure (top to bottom)
1. **Top row**: Company name (left, semi-bold, `IBM Plex Sans`) + RSU badge (green) + refresher badge (blue `↺`) on the right, if applicable
2. **Domain line**: muted, monospace, small font
3. **Notes**: slightly brighter, readable font size, `line-height: 1.55`
4. **Tag chips**: row of small pill chips, clickable to activate tag filter

### Footer
A muted disclaimer box explaining data sources (Levels.fyi, AmbitionBox, Glassdoor 2024–25) and caveats about comp ranges being approximate.

---

## Styling

### Theme
| Token | Value |
|---|---|
| Page background | `#0f1117` |
| Card background | `#131620` |
| Header background | `#0d0f18` |
| Border (default) | `#1e2231` |
| Text primary | `#f1f5f9` |
| Text secondary | `#94a3b8` |
| Text muted | `#64748b` |
| Text faint | `#4a5568` |

### Typography
Import both from Google Fonts via `@import` inside a JSX `<style>` tag:
- `IBM Plex Mono` (weights 300, 400, 500, 600) — all meta/code/domain text
- `IBM Plex Sans` (weights 300, 400, 500, 600, 700) — headings and company names

### Tier Color Palette
| Tier | Border | Text | Badge bg |
|---|---|---|---|
| T1 | `#f59e0b` | `#b45309` | `#fef3c7` |
| T2 | `#22c55e` | `#15803d` | `#dcfce7` |
| T3 | `#3b82f6` | `#1d4ed8` | `#dbeafe` |
| T4 | `#a855f7` | `#7e22ce` | `#f3e8ff` |

### Interactions
- Cards: hover lift `translateY(-1px)`, border lightens to `#4a5568`, background shifts to `#1a1d27`, `box-shadow: 0 4px 20px rgba(0,0,0,0.3)`
- All transitions: `0.15s ease`
- Tag chips: `cursor: pointer`, `opacity: 0.85` on hover, active state uses `#1e2535` background and `#94a3b8` text
- Custom scrollbar: `width: 6px`, track `#1a1d27`, thumb `#3b4155`, `border-radius: 3px`

---

## Technical Constraints

- Single `.jsx` file, default export, no required props
- Use only `useState` and `useMemo` from React — no other hooks
- No external UI libraries (no Tailwind, no shadcn, no MUI)
- No `localStorage`, `sessionStorage`, or any browser storage APIs — all state in memory
- Inline styles only (no CSS modules, no styled-components)
- Google Fonts loaded via `@import` inside a `<style>` tag injected in JSX
- All filter interactions update state without page reload

---

## Suggested Extensions

These can be added as follow-on instructions to Claude Code:

- `Add a "Shortlist" toggle per card, persisted in React state, with a Shortlist-only filter chip`
- `Add sort options: by tier, alphabetical, or by estimated comp ceiling`
- `Add a comp range slider filter (e.g. ₹30–80 LPA) parsed from the notes field`
- `Add a "Copy list" button that exports the filtered company names to clipboard`
- `Add a Levels.fyi deep-link on each card that opens the company's Bangalore salary page`
