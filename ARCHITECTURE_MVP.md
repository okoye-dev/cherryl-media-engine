# Cherryl Media — Product MVP Flow & Architecture

> [!NOTE]
> For the comprehensive, up-to-date production specification, please refer to the primary document:
> [**Cherryl Media — Product Architecture & System Blueprint (`ARCHITECTURE_PRODUCT.md`)**](./ARCHITECTURE_PRODUCT.md)

---

## 1. System Architecture Overview

```mermaid
flowchart TD
    subgraph Frontend ["Frontend (Next.js 15)"]
        PublicSite["Public Reader Site\n(/, /article/:slug, /category/:slug)"]
        AdminStudio["Admin Studio Newsroom\n(/admin, /admin/editor, /admin/sections, /admin/settings)"]
    end

    subgraph Backend ["Backend Engine (Go Media Engine)"]
        AuthService["Auth Service (JWT / Supabase ECC)"]
        ArticleService["Article & Editorial Service"]
        SectionService["Curated Sections Service"]
        SettingsService["Site Settings Service"]
        UploadService["S3 Presign Upload Service"]
    end

    subgraph Storage ["Infrastructure & Storage"]
        DB[("Supabase PostgreSQL\n(Articles, Curated Sections, Settings, Users)")]
        R2[("Cloudflare R2 (S3 Compatible)\nPublic Images & Videos Bucket")]
    end

    PublicSite -->|GET Published Articles & Sections| ArticleService
    AdminStudio -->|Admin Login / Session| AuthService
    AdminStudio -->|Autosave & Publish| ArticleService
    AdminStudio -->|Curate Layout Slots| SectionService
    AdminStudio -->|Update Newsroom Copy| SettingsService
    AdminStudio -->|Request Presigned Upload URL| UploadService
    UploadService -.->|Direct Client Upload (PUT)| R2
    R2 -.->|Public CDN URL| AdminStudio

    AuthService --> DB
    ArticleService --> DB
    SectionService --> DB
    SettingsService --> DB
```

---

## 2. Core Flows Summary

### A. Admin Authentication Flow
- Editor logs in at `/admin/login`.
- Go backend validates credentials via local HMAC or Supabase ECC JWKS.
- Issues JWT token stored in browser `localStorage` and injected via `ApiClient`.

### B. Story Editorial & Publishing Flow
- Writer authors story in the block-based editor (`/admin/editor`).
- Debounced slug uniqueness checking against `/api/articles/check-slug`.
- Typeahead tag search via `/api/tags`.
- Pre-publish validation modal confirms required fields.
- Post is saved and published directly to Supabase PostgreSQL.

### C. Homepage Curation & Settings Flows
- Curators visually assign stories to homepage slots at `/admin/sections`.
- Editors customize breaking news ticker notices and masthead copy at `/admin/settings`.

### D. Image & Video Upload Pipeline
- Frontend requests presigned S3/R2 upload URL from `POST /api/uploads/presign`.
- Browser uploads binary directly to Cloudflare R2 bucket.
- Storage keys (`media/articles/...`) are stored in PostgreSQL; `MediaResolver` hydrates full public CDN URLs upon retrieval.
