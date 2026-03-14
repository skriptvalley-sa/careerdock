
import { useState, useMemo } from "react";

const companies = [
  // Tier 1: FAANG-level / Top Comp
  { name: "Google", tier: 1, domain: "Search / AI / Cloud", rsu: true, refresher: true, notes: "L4–L6 most active, strong infra/ML hiring. ₹40–90 LPA for SDE2/SDE3.", tags: ["AI/ML", "Cloud", "Infra"] },
  { name: "Microsoft", tier: 1, domain: "Cloud / Enterprise / Dev Tools", rsu: true, refresher: true, notes: "Azure, GitHub, LinkedIn under one roof. Large Bangalore R&D. ₹35–80 LPA.", tags: ["Cloud", "Infra", "Dev Tools"] },
  { name: "Amazon / AWS", tier: 1, domain: "E-Commerce / Cloud", rsu: true, refresher: true, notes: "Front-loaded RSU structure. SDE2/SDE3 comp ₹40–85 LPA. AWS roles pay premium.", tags: ["Cloud", "Infra", "Distributed"] },
  { name: "Apple", tier: 1, domain: "Consumer / Hardware-Software", rsu: true, refresher: true, notes: "Selective hiring, premium pay. Siri, Maps, iCloud backend roles. ₹45–90 LPA.", tags: ["Infra", "Embedded"] },
  { name: "Meta", tier: 1, domain: "Social / AI", rsu: true, refresher: true, notes: "Smaller Blr presence, mostly Hyderabad/Gurugram. Very high pay for SWE.", tags: ["AI/ML", "Infra"] },

  // Tier 2: Strong MNCs with RSUs
  { name: "Adobe", tier: 2, domain: "Creative / Digital Experience", rsu: true, refresher: true, notes: "Large Bangalore R&D. Photoshop, AEM, Firefly. ₹30–65 LPA for senior SWEs.", tags: ["Cloud", "Platform"] },
  { name: "Salesforce", tier: 2, domain: "CRM / Cloud SaaS", rsu: true, refresher: true, notes: "Hyperforce (cloud infra) team is growing. Good WLB + solid equity.", tags: ["Cloud", "SaaS"] },
  { name: "Cisco", tier: 2, domain: "Networking / Security", rsu: true, refresher: true, notes: "Massive Bangalore engineering base. Splunk integration expanding cloud security roles.", tags: ["Networking", "Security"] },
  { name: "Intel", tier: 2, domain: "Semiconductors / Software", rsu: true, refresher: true, notes: "Chip software, compilers, platform engineering. ₹30–60 LPA.", tags: ["Embedded", "Infra"] },
  { name: "Qualcomm", tier: 2, domain: "Semiconductors / Mobile", rsu: true, refresher: true, notes: "Strong comp for firmware/SoC roles. Annual refreshers standard.", tags: ["Embedded"] },
  { name: "Oracle", tier: 2, domain: "Database / Cloud / ERP", rsu: true, refresher: true, notes: "OCI (Oracle Cloud) team in Bangalore growing rapidly. Decent RSU grants.", tags: ["Cloud", "Database"] },
  { name: "SAP Labs", tier: 2, domain: "Enterprise ERP / Business AI", rsu: true, refresher: true, notes: "Largest R&D hub outside Germany. HANA, AI business apps. ₹25–55 LPA.", tags: ["SaaS", "Platform"] },
  { name: "Intuit", tier: 2, domain: "FinTech SaaS", rsu: true, refresher: true, notes: "TurboTax, QuickBooks. Strong comp + good WLB. ₹28–60 LPA SWE-III+.", tags: ["FinTech", "SaaS"] },
  { name: "Atlassian", tier: 2, domain: "Dev Tools / Collaboration", rsu: true, refresher: true, notes: "Jira, Confluence, Bitbucket. Remote-friendly culture, solid equity.", tags: ["Dev Tools", "Platform"] },
  { name: "LinkedIn", tier: 2, domain: "Professional Networking / AI", rsu: true, refresher: true, notes: "Microsoft subsidiary. Backend, AI-matching, data infra. ₹30–65 LPA.", tags: ["AI/ML", "Platform"] },
  { name: "Uber", tier: 2, domain: "Mobility / Logistics", rsu: true, refresher: true, notes: "Large Bangalore engineering org. Platform/Infra/ML roles active.", tags: ["Distributed", "Platform"] },
  { name: "Walmart Global Tech", tier: 2, domain: "Retail Tech / E-Commerce", rsu: true, refresher: true, notes: "Parent also owns Flipkart + PhonePe. Global tech roles, strong RSUs.", tags: ["Platform", "Distributed"] },
  { name: "PayPal", tier: 2, domain: "FinTech / Payments", rsu: true, refresher: true, notes: "Good mid-senior SWE comp. Payments platform, fraud detection.", tags: ["FinTech", "Platform"] },
  { name: "Visa", tier: 2, domain: "FinTech / Payments Infrastructure", rsu: true, refresher: true, notes: "Bangalore tech hub. Core payment rails engineering. Stable + strong equity.", tags: ["FinTech", "Infra"] },
  { name: "Mastercard", tier: 2, domain: "FinTech / Payments", rsu: true, refresher: true, notes: "Labs + tech center in Bangalore. Good RSU component for senior SWEs.", tags: ["FinTech"] },
  { name: "American Express", tier: 2, domain: "FinTech / Consumer Finance", rsu: true, refresher: true, notes: "GCC in Bangalore with strong tech roles. Equity grants included.", tags: ["FinTech", "Platform"] },
  { name: "Broadcom", tier: 2, domain: "Semiconductors / VMware", rsu: true, refresher: true, notes: "Includes former VMware teams. Highest median comp on Levels.fyi for Blr. ₹60–120 LPA.", tags: ["Infra", "Networking"] },
  { name: "ServiceNow", tier: 2, domain: "Enterprise Workflow / Cloud", rsu: true, refresher: true, notes: "Fast-growing Bangalore team. Strong refreshers, good WLB.", tags: ["SaaS", "Cloud"] },
  { name: "Workday", tier: 2, domain: "HR / Finance SaaS", rsu: true, refresher: true, notes: "Bangalore tech center. Solid mid-senior SWE comp with equity.", tags: ["SaaS"] },
  { name: "Palo Alto Networks", tier: 2, domain: "Cybersecurity", rsu: true, refresher: true, notes: "Expanding Bangalore presence. Cloud security, SIEM, SOAR roles.", tags: ["Security", "Cloud"] },
  { name: "CrowdStrike", tier: 2, domain: "Cybersecurity / EDR", rsu: true, refresher: true, notes: "Good equity, fast-growing. Platform and detection engineering roles.", tags: ["Security", "Platform"] },
  { name: "Zscaler", tier: 2, domain: "Cloud Security / Zero Trust", rsu: true, refresher: true, notes: "Bangalore engineering team. Cloud networking + security stack.", tags: ["Security", "Cloud", "Networking"] },

  // Tier 3: Good product MNCs — solid RSU, slightly lower ceiling
  { name: "MongoDB", tier: 3, domain: "Database / Cloud", rsu: true, refresher: true, notes: "Atlas cloud team active. Smaller Bangalore office but good equity.", tags: ["Database", "Cloud"] },
  { name: "Elastic", tier: 3, domain: "Search / Observability", rsu: true, refresher: false, notes: "Search + SIEM. Refreshers less standard. Fully remote-friendly.", tags: ["Platform", "Infra"] },
  { name: "Confluent", tier: 3, domain: "Data Streaming / Kafka", rsu: true, refresher: true, notes: "Kafka-based platform. Good equity, active Bangalore hiring.", tags: ["Distributed", "Platform"] },
  { name: "Databricks", tier: 3, domain: "Data / AI Platform", rsu: true, refresher: true, notes: "Pre-IPO equity can be very valuable. Strong engineering culture.", tags: ["AI/ML", "Data", "Platform"] },
  { name: "Cloudflare", tier: 3, domain: "CDN / Edge / Security", rsu: true, refresher: false, notes: "Edge compute and networking. Growing India team.", tags: ["Networking", "Security", "Infra"] },
  { name: "Okta", tier: 3, domain: "Identity / Security", rsu: true, refresher: false, notes: "Auth0 merged. IAM platform engineering roles.", tags: ["Security", "Platform"] },
  { name: "Twilio", tier: 3, domain: "Communications API / SaaS", rsu: true, refresher: false, notes: "Platform and infra roles. Post-restructure hiring is selective.", tags: ["Platform", "SaaS"] },
  { name: "Nutanix", tier: 3, domain: "HCI / Cloud Infra", rsu: true, refresher: true, notes: "Hyper-converged infra. Good comp for SRE/infra engineers. ₹25–55 LPA.", tags: ["Infra", "Cloud"] },
  { name: "Pure Storage", tier: 3, domain: "Storage / Cloud", rsu: true, refresher: true, notes: "Storage platform + cloud. Solid Bangalore team.", tags: ["Infra", "Storage"] },
  { name: "NetApp", tier: 3, domain: "Storage / Cloud Data", rsu: true, refresher: true, notes: "Long-standing Bangalore R&D. Good equity for senior SWEs.", tags: ["Infra", "Storage", "Cloud"] },
  { name: "Rubrik", tier: 3, domain: "Data Security / Backup Cloud", rsu: true, refresher: true, notes: "Post-IPO RSUs now public. Strong Bangalore engineering team.", tags: ["Security", "Cloud", "Infra"] },
  { name: "Cohesity", tier: 3, domain: "Data Management / AI", rsu: true, refresher: false, notes: "Pre-IPO stage. Strong equity potential for infra engineers.", tags: ["Infra", "AI/ML"] },
  { name: "JFrog", tier: 3, domain: "DevOps / Artifact Management", rsu: true, refresher: false, notes: "Artifactory, Xray. Public company. Bangalore DevOps platform team.", tags: ["Dev Tools", "Platform"] },
  { name: "Dynatrace", tier: 3, domain: "Observability / APM", rsu: true, refresher: false, notes: "Monitoring + AI ops. Growing Bangalore presence.", tags: ["Platform", "AI/ML"] },
  { name: "New Relic", tier: 3, domain: "Observability", rsu: true, refresher: false, notes: "SaaS observability. Bangalore backend team.", tags: ["Platform"] },
  { name: "Marvell Technology", tier: 3, domain: "Semiconductors / Networking", rsu: true, refresher: true, notes: "Chip design + firmware. Good comp for embedded/networking SWEs.", tags: ["Networking", "Embedded"] },
  { name: "Arista Networks", tier: 3, domain: "Network OS / Cloud Networking", rsu: true, refresher: true, notes: "EOS (network OS) engineering. Strong comp for networking SWEs.", tags: ["Networking", "Infra"] },
  { name: "Juniper Networks", tier: 3, domain: "Networking", rsu: true, refresher: true, notes: "AI-driven networks. Bangalore R&D, Apstra/Mist product teams.", tags: ["Networking", "AI/ML"] },
  { name: "Synopsys", tier: 3, domain: "EDA / Chip Design", rsu: true, refresher: true, notes: "EDA tools. Strong base for chip/embedded SW engineers.", tags: ["Embedded", "Platform"] },
  { name: "Cadence Design Systems", tier: 3, domain: "EDA / SoC Design", rsu: true, refresher: true, notes: "EDA + verification. Good equity + WLB. ₹25–55 LPA.", tags: ["Embedded", "Platform"] },
  { name: "Texas Instruments", tier: 3, domain: "Semiconductors / Embedded", rsu: true, refresher: true, notes: "Embedded software. Stable company, long vesting cycles, strong benefits.", tags: ["Embedded"] },
  { name: "ARM (SoftBank)", tier: 3, domain: "Chip Architecture / IP", rsu: true, refresher: true, notes: "Architecture, firmware, toolchain engineering. Post-IPO RSUs.", tags: ["Embedded", "Platform"] },
  { name: "Western Digital", tier: 3, domain: "Storage / Embedded", rsu: true, refresher: false, notes: "NAND/SSD firmware. Stable MNC with equity.", tags: ["Storage", "Embedded"] },
  { name: "Seagate", tier: 3, domain: "Storage Hardware / Cloud", rsu: true, refresher: false, notes: "Firmware and cloud data infra teams in Bangalore.", tags: ["Storage", "Infra"] },
  { name: "Micron Technology", tier: 3, domain: "Memory / Embedded", rsu: true, refresher: true, notes: "Memory chip SW. Growing Bangalore R&D team.", tags: ["Embedded"] },
  { name: "HP / HPE", tier: 3, domain: "Enterprise Hardware / Cloud", rsu: true, refresher: false, notes: "HPE more relevant (Greenlake cloud platform). RSUs for senior roles.", tags: ["Cloud", "Infra"] },
  { name: "Dell Technologies", tier: 3, domain: "Enterprise / Storage / Cloud", rsu: true, refresher: false, notes: "VMware heritage teams now under Broadcom. Dell remains active in Blr.", tags: ["Infra", "Cloud"] },
  { name: "Honeywell", tier: 3, domain: "Industrial Tech / IoT", rsu: true, refresher: false, notes: "HCE (Honeywell Connected Enterprise) GCC in Bangalore. Stable.", tags: ["Platform", "Embedded"] },
  { name: "Bosch", tier: 3, domain: "Automotive / IoT / Embedded", rsu: false, refresher: false, notes: "RBEI (Bosch Global Software Tech). Very large Bangalore presence. ESOPs for leadership.", tags: ["Embedded", "Automotive"] },
  { name: "Samsung R&D", tier: 3, domain: "Consumer Electronics / AI / Mobile", rsu: true, refresher: false, notes: "SRIB (Samsung Research India). Good pay, large team.", tags: ["AI/ML", "Embedded"] },
  { name: "Siemens EDA (Mentor)", tier: 3, domain: "EDA / Industrial", rsu: true, refresher: false, notes: "EDA toolchain. Siemens Digital Industries team.", tags: ["Embedded", "Platform"] },

  // Tier 4: Indian product companies with RSU/ESOP
  { name: "Flipkart", tier: 4, domain: "E-Commerce / Platform", rsu: true, refresher: false, notes: "Walmart-owned. ESOP/RSU for senior roles. Strong Blr platform team.", tags: ["Platform", "Distributed"] },
  { name: "PhonePe", tier: 4, domain: "FinTech / Payments", rsu: true, refresher: false, notes: "Walmart-backed. IPO anticipated. Strong pay for senior SWEs.", tags: ["FinTech", "Platform"] },
  { name: "Swiggy", tier: 4, domain: "Food Tech / Logistics", rsu: true, refresher: false, notes: "Listed company now. RSUs granted. Platform/infra team solid.", tags: ["Platform", "Distributed"] },
  { name: "Zomato / Blinkit", tier: 4, domain: "Food Tech / Quick Commerce", rsu: true, refresher: false, notes: "Listed (NSE: ZOMATO). RSUs. Strong data/platform engineering team.", tags: ["Platform", "AI/ML"] },
  { name: "Razorpay", tier: 4, domain: "FinTech / Payments", rsu: true, refresher: false, notes: "Pre-IPO. ESOP potential high. Strong Bangalore engineering base.", tags: ["FinTech", "Platform"] },
  { name: "Freshworks", tier: 4, domain: "CRM / SaaS", rsu: true, refresher: false, notes: "Listed (NASDAQ: FRSH). RSUs. Mature SaaS engineering culture.", tags: ["SaaS", "Platform"] },
  { name: "CRED", tier: 4, domain: "FinTech / Lifestyle", rsu: true, refresher: false, notes: "Unicorn. Pre-IPO ESOPs. Strong tech culture, selective hiring.", tags: ["FinTech", "Platform"] },
  { name: "Meesho", tier: 4, domain: "Social Commerce", rsu: true, refresher: false, notes: "Pre-IPO. ESOP upside. Data + platform engineering.", tags: ["Platform"] },
  { name: "Groww", tier: 4, domain: "FinTech / Stock Broking", rsu: true, refresher: false, notes: "Pre-IPO unicorn. Good pay + equity. Engineering-first culture.", tags: ["FinTech", "Platform"] },
  { name: "Zepto", tier: 4, domain: "Quick Commerce", rsu: true, refresher: false, notes: "Hypergrowth phase. Pre-IPO equity. Infra/platform hiring.", tags: ["Platform", "Distributed"] },
  { name: "InMobi", tier: 4, domain: "AdTech / AI", rsu: true, refresher: false, notes: "Global AdTech, Bangalore HQ. Pre-IPO. DSP + ML engineering.", tags: ["AI/ML", "Platform"] },
  { name: "Browserstack", tier: 4, domain: "Dev Tools / Testing SaaS", rsu: true, refresher: false, notes: "Profitable bootstrapped company. ESOPs. Strong SWE comp.", tags: ["Dev Tools", "SaaS"] },
  { name: "Clevertap", tier: 4, domain: "Customer Engagement / SaaS", rsu: true, refresher: false, notes: "Growth SaaS. Pre-IPO ESOPs. Platform engineering roles.", tags: ["SaaS", "Platform"] },
];

const TIER_LABELS = {
  1: "Tier 1 — FAANG & Peak Comp",
  2: "Tier 2 — Top MNCs with Strong RSUs",
  3: "Tier 3 — Solid MNCs / Good Equity",
  4: "Tier 4 — Indian Product Companies",
};

const TIER_COLORS = {
  1: { bg: "#fff8e1", border: "#f59e0b", badge: "#b45309", badgeBg: "#fef3c7" },
  2: { bg: "#f0fdf4", border: "#22c55e", badge: "#15803d", badgeBg: "#dcfce7" },
  3: { bg: "#eff6ff", border: "#3b82f6", badge: "#1d4ed8", badgeBg: "#dbeafe" },
  4: { bg: "#fdf4ff", border: "#a855f7", badge: "#7e22ce", badgeBg: "#f3e8ff" },
};

const ALL_TAGS = [...new Set(companies.flatMap(c => c.tags))].sort();

export default function App() {
  const [search, setSearch] = useState("");
  const [selectedTiers, setSelectedTiers] = useState([1, 2, 3, 4]);
  const [selectedTags, setSelectedTags] = useState([]);
  const [rsuOnly, setRsuOnly] = useState(false);
  const [refresherOnly, setRefresherOnly] = useState(false);

  const toggleTier = (t) =>
    setSelectedTiers(prev => prev.includes(t) ? prev.filter(x => x !== t) : [...prev, t]);
  const toggleTag = (tag) =>
    setSelectedTags(prev => prev.includes(tag) ? prev.filter(x => x !== tag) : [...prev, tag]);

  const filtered = useMemo(() => {
    return companies.filter(c => {
      if (!selectedTiers.includes(c.tier)) return false;
      if (rsuOnly && !c.rsu) return false;
      if (refresherOnly && !c.refresher) return false;
      if (selectedTags.length > 0 && !selectedTags.some(t => c.tags.includes(t))) return false;
      if (search && !c.name.toLowerCase().includes(search.toLowerCase()) &&
          !c.domain.toLowerCase().includes(search.toLowerCase()) &&
          !c.notes.toLowerCase().includes(search.toLowerCase())) return false;
      return true;
    });
  }, [search, selectedTiers, selectedTags, rsuOnly, refresherOnly]);

  const byTier = useMemo(() => {
    const map = {};
    filtered.forEach(c => {
      if (!map[c.tier]) map[c.tier] = [];
      map[c.tier].push(c);
    });
    return map;
  }, [filtered]);

  return (
    <div style={{
      minHeight: "100vh",
      background: "#0f1117",
      color: "#e2e8f0",
      fontFamily: "'IBM Plex Mono', 'Fira Code', 'Cascadia Code', monospace",
      padding: "0",
    }}>
      <style>{`
        @import url('https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@300;400;500;600&family=IBM+Plex+Sans:wght@300;400;500;600;700&display=swap');
        * { box-sizing: border-box; margin: 0; padding: 0; }
        ::-webkit-scrollbar { width: 6px; }
        ::-webkit-scrollbar-track { background: #1a1d27; }
        ::-webkit-scrollbar-thumb { background: #3b4155; border-radius: 3px; }
        .card:hover { border-color: #4a5568 !important; background: #1a1d27 !important; transform: translateY(-1px); box-shadow: 0 4px 20px rgba(0,0,0,0.3); }
        .card { transition: all 0.15s ease; }
        .tag-btn:hover { opacity: 0.85; }
        .filter-chip { cursor: pointer; user-select: none; transition: all 0.15s; }
        .filter-chip:hover { opacity: 0.8; }
        input::placeholder { color: #4a5568; }
        input:focus { outline: none; border-color: #4a90e2 !important; }
      `}</style>

      {/* Header */}
      <div style={{
        borderBottom: "1px solid #1e2231",
        padding: "28px 32px 20px",
        background: "#0d0f18",
        position: "sticky", top: 0, zIndex: 10,
      }}>
        <div style={{ maxWidth: 1100, margin: "0 auto" }}>
          <div style={{ display: "flex", alignItems: "baseline", gap: 16, marginBottom: 16 }}>
            <h1 style={{
              fontSize: 22,
              fontFamily: "'IBM Plex Sans', sans-serif",
              fontWeight: 700,
              letterSpacing: "-0.5px",
              color: "#f1f5f9",
            }}>
              🏢 Bangalore Tech Companies
            </h1>
            <span style={{
              fontSize: 12,
              fontFamily: "'IBM Plex Mono', monospace",
              color: "#64748b",
              fontWeight: 400,
            }}>
              {filtered.length} of {companies.length} companies · ~5 YOE · RSU + Refreshers
            </span>
          </div>

          {/* Search + Toggles */}
          <div style={{ display: "flex", gap: 12, flexWrap: "wrap", alignItems: "center", marginBottom: 14 }}>
            <input
              value={search}
              onChange={e => setSearch(e.target.value)}
              placeholder="Search companies, domains, keywords..."
              style={{
                background: "#1a1d27",
                border: "1px solid #2a2f42",
                borderRadius: 6,
                padding: "8px 14px",
                color: "#e2e8f0",
                fontSize: 13,
                width: 280,
                fontFamily: "'IBM Plex Mono', monospace",
              }}
            />
            <label style={{ display: "flex", alignItems: "center", gap: 7, cursor: "pointer", fontSize: 12, color: "#94a3b8" }}>
              <input type="checkbox" checked={rsuOnly} onChange={e => setRsuOnly(e.target.checked)}
                style={{ accentColor: "#4a90e2", width: 14, height: 14 }} />
              RSU only
            </label>
            <label style={{ display: "flex", alignItems: "center", gap: 7, cursor: "pointer", fontSize: 12, color: "#94a3b8" }}>
              <input type="checkbox" checked={refresherOnly} onChange={e => setRefresherOnly(e.target.checked)}
                style={{ accentColor: "#4a90e2", width: 14, height: 14 }} />
              With Refreshers
            </label>
          </div>

          {/* Tier filters */}
          <div style={{ display: "flex", gap: 8, flexWrap: "wrap", marginBottom: 10 }}>
            {[1, 2, 3, 4].map(t => {
              const col = TIER_COLORS[t];
              const active = selectedTiers.includes(t);
              return (
                <div key={t}
                  className="filter-chip"
                  onClick={() => toggleTier(t)}
                  style={{
                    padding: "4px 12px",
                    borderRadius: 20,
                    fontSize: 11,
                    fontWeight: 500,
                    border: `1px solid ${active ? col.border : "#2a2f42"}`,
                    color: active ? col.badge : "#64748b",
                    background: active ? col.badgeBg + "22" : "transparent",
                    cursor: "pointer",
                  }}>
                  T{t} · {byTier[t]?.length ?? 0}
                </div>
              );
            })}
            <div style={{ width: 1, background: "#2a2f42", margin: "0 4px" }} />
            {ALL_TAGS.map(tag => {
              const active = selectedTags.includes(tag);
              return (
                <div key={tag}
                  className="filter-chip"
                  onClick={() => toggleTag(tag)}
                  style={{
                    padding: "4px 10px",
                    borderRadius: 20,
                    fontSize: 11,
                    border: `1px solid ${active ? "#64748b" : "#1e2231"}`,
                    color: active ? "#cbd5e1" : "#4a5568",
                    background: active ? "#1e2535" : "transparent",
                    cursor: "pointer",
                  }}>
                  {tag}
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Content */}
      <div style={{ maxWidth: 1100, margin: "0 auto", padding: "24px 32px 48px" }}>
        {[1, 2, 3, 4].map(tier => {
          const list = byTier[tier];
          if (!list || list.length === 0) return null;
          const col = TIER_COLORS[tier];
          return (
            <div key={tier} style={{ marginBottom: 36 }}>
              <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 14 }}>
                <div style={{ height: 1, width: 20, background: col.border, opacity: 0.6 }} />
                <span style={{ fontSize: 11, fontWeight: 600, color: col.badge, letterSpacing: "0.08em", textTransform: "uppercase" }}>
                  {TIER_LABELS[tier]}
                </span>
                <div style={{ height: 1, flex: 1, background: "#1e2231" }} />
                <span style={{ fontSize: 11, color: "#4a5568" }}>{list.length}</span>
              </div>

              <div style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fill, minmax(300px, 1fr))",
                gap: 10,
              }}>
                {list.map(c => (
                  <div key={c.name}
                    className="card"
                    style={{
                      background: "#131620",
                      border: `1px solid #1e2231`,
                      borderRadius: 8,
                      padding: "14px 16px",
                      cursor: "default",
                    }}>
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 6 }}>
                      <span style={{ fontSize: 14, fontWeight: 600, color: "#f1f5f9", fontFamily: "'IBM Plex Sans', sans-serif" }}>
                        {c.name}
                      </span>
                      <div style={{ display: "flex", gap: 5 }}>
                        {c.rsu && (
                          <span style={{ fontSize: 10, padding: "2px 7px", borderRadius: 4, background: "#14291a", color: "#4ade80", border: "1px solid #166534", fontWeight: 500 }}>
                            RSU
                          </span>
                        )}
                        {c.refresher && (
                          <span style={{ fontSize: 10, padding: "2px 7px", borderRadius: 4, background: "#1a2440", color: "#60a5fa", border: "1px solid #1d4ed8", fontWeight: 500 }}>
                            ↺
                          </span>
                        )}
                      </div>
                    </div>
                    <div style={{ fontSize: 11, color: "#64748b", marginBottom: 7, fontFamily: "'IBM Plex Mono', monospace" }}>
                      {c.domain}
                    </div>
                    <div style={{ fontSize: 11.5, color: "#94a3b8", lineHeight: 1.55, marginBottom: 8 }}>
                      {c.notes}
                    </div>
                    <div style={{ display: "flex", gap: 5, flexWrap: "wrap" }}>
                      {c.tags.map(tag => (
                        <span key={tag}
                          className="tag-btn"
                          onClick={() => toggleTag(tag)}
                          style={{
                            fontSize: 10,
                            padding: "2px 8px",
                            borderRadius: 20,
                            background: selectedTags.includes(tag) ? "#1e2535" : "#1a1d27",
                            color: selectedTags.includes(tag) ? "#94a3b8" : "#4a5568",
                            border: `1px solid ${selectedTags.includes(tag) ? "#3b4155" : "#1e2231"}`,
                            cursor: "pointer",
                          }}>
                          {tag}
                        </span>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          );
        })}

        {filtered.length === 0 && (
          <div style={{ textAlign: "center", color: "#4a5568", padding: 60, fontSize: 13 }}>
            No companies match the current filters.
          </div>
        )}

        <div style={{ marginTop: 32, padding: "14px 20px", background: "#0d0f18", borderRadius: 8, border: "1px solid #1e2231" }}>
          <p style={{ fontSize: 11, color: "#4a5568", lineHeight: 1.7 }}>
            <span style={{ color: "#64748b" }}>Note:</span> Comp ranges are approximate (~5 YOE) and sourced from Levels.fyi, AmbitionBox, Glassdoor (2024–25).
            RSU/refresher details may vary by role, level, and negotiation. Always verify on Levels.fyi before interviews.
            ↺ = refreshers confirmed common. Indian cos under T4 show ESOPs, not listed RSUs unless noted.
          </p>
        </div>
      </div>
    </div>
  );
}
