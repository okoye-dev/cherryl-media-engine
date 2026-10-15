# Cherryl Media: Product, Automation & Growth Roadmap

## 1. Vision & Core Principles
Build a lightweight, repeatable modern publishing operation for a small team:
- **Discover**: Monitor trusted public sources automatically.
- **Select**: Human editorial decision on what deserves coverage.
- **Generate**: AI-assisted, source-grounded draft with verified facts and attribution.
- **Review**: Human editorial verification, editing, and approval (AI never publishes directly).
- **Distribute**: One canonical story adapted to Website, Facebook, X, LinkedIn, Newsletter.
- **Grow**: Commercialize via direct audience, media campaigns, and sponsorships.

---

## 2. Editorial Pillars & Channels

| Pillar | Example Content | Automation Level | Primary Channels |
| :--- | :--- | :--- | :--- |
| **Nigerian Current Affairs** | Confirmed announcements, policy changes, major developments | Medium | Website, Facebook, X |
| **Entertainment & Internet Culture** | Releases, Nollywood/music, trending explainers | Med–High | Facebook, X, TikTok, Website |
| **Practical Nigerian Updates** | Education, transport, electricity, public services | High | Website, Facebook, Newsletter |
| **Business & Institutions** | Corporate announcements, fintech, industry reports | Med–High | Website, LinkedIn, X, Newsletter |
| **Stories & Romance** | Modern love, personal essays, relationship chronicles | Curated | Website, Socials, Weekend Edition |

---

## 3. One Story, Multichannel Products
Backend stores **1 Canonical Story**, then adapts:
- **Website**: Full searchable article with Medium-style blocks (drop-cap, pull-quotes, inline photos, source links).
- **Facebook**: Strong opening hook, social image, readable explanation, link.
- **X (Twitter)**: Concise 280-char update with key fact + thread where warranted.
- **LinkedIn**: Institutional/economic angle: who is affected and why leaders should care.
- **Newsletter**: Curated morning/weekend digest summary.

---

## 4. Rollout Phases

- **Weeks 1–2 (Phase 1: Foundation)**: Audit WordPress archive, define PostgreSQL schema, configure media storage.
- **Weeks 3–4 (Phase 2: Website & Migration)**: Launch Next.js site + custom `/admin` editor. Stream images via Go migration script.
- **Weeks 5–6 (Phase 3: Editorial Pipeline)**: Go ingestion workers (RSS/APIs), story deduplication, candidate inbox.
- **Weeks 7–8 (Phase 4: Distribution)**: Automated social adapters (FB, X, LinkedIn, Newsletter) with human approval gates.
- **Weeks 9–12 (Phase 5: Commercialization)**: Media kit, sponsorship packages, audience growth measurement.

---

## 5. What NOT to Build
- No complex multi-tier microservices on day 1 (monolithic Go binary + Next.js).
- No autonomous AI publishing (human review always required).
- No generic commenting/forums.
- No heavy expensive CMS subscriptions (Sanity eliminated; self-hosted Next.js admin + Postgres + direct uploads).
