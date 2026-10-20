# Cherryl Media — Legacy Category Migration Mapping

> **IMPORTANT**: This document defines the definitive migration rules from the legacy WordPress website (`cherrylmedia.com`) to the new Cherryl Media platform.
> The new platform strictly maintains **10 official editorial pillars**. All legacy articles must be mapped either to one of the 10 pillars or routed to the designated catch-all category.

---

## 1. The 10 Official Target Pillars

The new site exclusively supports these 10 categories for its primary navigation, archives, and taxonomy:

| Target Pillar Slug | Display Name | Scope / Focus |
|---|---|---|
| `state-affairs-environs` | **STATE AFFAIRS & ENVIRONS** | Sub-national policies, state budgets, environment, regional infrastructure. |
| `newsbeat` | **NEWSBEAT** | Breaking news, national alerts, investigative reports, security dispatches. |
| `brands-smes` | **BRANDS & SMEs** | Enterprise growth, SME focus, market trends, consumer brands, retail. |
| `opinions` | **OPINIONS** | Editorial columns, policy analysis, guest essays, cultural critique. |
| `sports` | **SPORTS** | Football (NPFL, Europe), athletics, combat sports, African sports. |
| `entertainment-lifestyle` | **ENTERTAINMENT & LIFESTYLE** | Nollywood, music scene, red carpet, creative arts, dining, lifestyle. |
| `education` | **EDUCATION** | Universities, polytechnics, WAEC/JAMB, scholarships, academic policy. |
| `thrills-tricks` | **THRILLS & TRICKS** | Everyday hacks, consumer tips, viral spectacles, curious finds. |
| `youth-corner` | **YOUTH CORNER** | Gen-Z voices, student creators, campus tech, youth startups. |
| `tour-hotels` | **TOUR & HOTELS** | Hospitality & stays, travel destinations, weekend getaways, heritage tourism. |

---

## 2. Complete Legacy Categories Inventory (20 Categories)

The legacy WordPress platform had 20 categories:

| Legacy Category Name | Legacy URL | Match Status | Target Action |
|---|---|---|---|
| `STATE AFFAIRS & ENVIRONS` | `https://cherrylmedia.com/category/state-affairs-environs/` | Exact Match | Map to `state-affairs-environs` |
| `NEWSBEAT` | `https://cherrylmedia.com/category/newsbeat/` | Exact Match | Map to `newsbeat` |
| `BRAND & SMEs` | `https://cherrylmedia.com/category/brand-smes/` | Minor Correction (Missing S) | **Map to `brands-smes` ("BRANDS & SMEs")** |
| `OPINION` | `https://cherrylmedia.com/category/opinion/` | Minor Correction (Missing S) | **Map to `opinions` ("OPINIONS")** |
| `SPORTS` | `https://cherrylmedia.com/category/sports/` | Exact Match | Map to `sports` |
| `ENTERTAINMENT & LIFESTYLE` | `https://cherrylmedia.com/category/entertainment-lifestyle-world/` | Slug Normalization | **Map to `entertainment-lifestyle`** |
| `EDUCATION` | `https://cherrylmedia.com/category/education/` | Exact Match | Map to `education` |
| `THRILLS & TRICKS` | `https://cherrylmedia.com/category/thrills-tricks/` | Exact Match | Map to `thrills-tricks` |
| `YOUTH CORNER` | `https://cherrylmedia.com/category/youth-corner/` | Exact Match | Map to `youth-corner` |
| `TOUR & HOTELS` | `https://cherrylmedia.com/category/tour-hotels/` | Exact Match | Map to `tour-hotels` |
| `ART & COMICS` | `https://cherrylmedia.com/category/art-comics/` | Non-Pillar | Route to Catch-All (or Subcategory of `entertainment-lifestyle`) |
| `BOOK & VIDEO REVIEW` | `https://cherrylmedia.com/category/book-video-review/` | Non-Pillar | Route to Catch-All (or Subcategory of `entertainment-lifestyle`) |
| `EVENTS` | `https://cherrylmedia.com/category/events/` | Non-Pillar | Route to Catch-All (or Subcategory of `entertainment-lifestyle`) |
| `FASHION & FRAGRANCE` | `https://cherrylmedia.com/category/fashion-fragrance/` | Non-Pillar | Route to Catch-All (or Subcategory of `entertainment-lifestyle`) |
| `FOOD & WINE` | `https://cherrylmedia.com/category/food-wine-2/` | Non-Pillar | Route to Catch-All (or Subcategory of `entertainment-lifestyle`) |
| `HEALTHY LIVING` | `https://cherrylmedia.com/category/healthy-living/` | Non-Pillar | Route to Catch-All (or Subcategory of `thrills-tricks` / `entertainment-lifestyle`) |
| `INTERVIEWS` | `https://cherrylmedia.com/category/interviews/` | Non-Pillar | Route to Catch-All (or Subcategory of `opinions` / `newsbeat`) |
| `PERSONALITY INTERVIEW` | `https://cherrylmedia.com/category/personality-interview/` | Non-Pillar | Route to Catch-All (or Subcategory of `opinions` / `newsbeat`) |
| `SCORE CARD & AWARD` | `https://cherrylmedia.com/category/score-card-award/` | Non-Pillar | Route to Catch-All (or Subcategory of `state-affairs-environs` / `brands-smes`) |
| `Uncategorized` | `https://cherrylmedia.com/category/uncategorized/` | Catch-all | Route to Catch-All |

---

## 3. Migration Rules & Routing Table

### Direct 1-to-1 Mapping Rules (With Corrections)

```
1. "STATE AFFAIRS & ENVIRONS"                 --> "state-affairs-environs" ("STATE AFFAIRS & ENVIRONS")
2. "NEWSBEAT"                                 --> "newsbeat" ("NEWSBEAT")
3. "BRAND & SMEs" (missing S)                 --> "brands-smes" ("BRANDS & SMEs")
4. "OPINION" (missing S)                      --> "opinions" ("OPINIONS")
5. "SPORTS"                                   --> "sports" ("SPORTS")
6. "ENTERTAINMENT & LIFESTYLE" (world slug)   --> "entertainment-lifestyle" ("ENTERTAINMENT & LIFESTYLE")
7. "EDUCATION"                                --> "education" ("EDUCATION")
8. "THRILLS & TRICKS"                         --> "thrills-tricks" ("THRILLS & TRICKS")
9. "YOUTH CORNER"                             --> "youth-corner" ("YOUTH CORNER")
10. "TOUR & HOTELS"                           --> "tour-hotels" ("TOUR & HOTELS")
```

### Catch-All Section Strategy for Non-Pillar Categories

For any legacy post belonging to categories outside the 10 official pillars, the migration script supports two configurable modes:

#### Option A: Preserve as Tag + Map to Closest Editorial Pillar (Recommended)
Add the legacy category name as a **Tag** (e.g. `#Art & Comics`, `#Fashion & Fragrance`, `#Interviews`) so that search and archive filtering works, while categorizing under the closest official pillar:
- `ART & COMICS`          -> Tag: `Art & Comics`, Category: `entertainment-lifestyle`
- `BOOK & VIDEO REVIEW`   -> Tag: `Book & Video Review`, Category: `entertainment-lifestyle`
- `EVENTS`                -> Tag: `Events`, Category: `entertainment-lifestyle`
- `FASHION & FRAGRANCE`   -> Tag: `Fashion & Fragrance`, Category: `entertainment-lifestyle`
- `FOOD & WINE`           -> Tag: `Food & Wine`, Category: `entertainment-lifestyle`
- `HEALTHY LIVING`        -> Tag: `Healthy Living`, Category: `thrills-tricks`
- `INTERVIEWS`            -> Tag: `Interviews`, Category: `opinions`
- `PERSONALITY INTERVIEW`  -> Tag: `Personality Interview`, Category: `opinions`
- `SCORE CARD & AWARD`    -> Tag: `Score Card & Award`, Category: `state-affairs-environs`
- `Uncategorized`         -> Tag: `Legacy Archive`, Category: `newsbeat`

#### Option B: Dedicated Standalone Catch-All Pillar
Route any unmapped post to a dedicated catch-all pillar:
- Slug: `special-features` / `archives`
- Display: **SPECIAL FEATURES & ARCHIVES**

---

## 4. Machine-Readable Migration JSON

This JSON configuration can be directly imported by Go / Node / Python migration scripts:

```json
{
  "target_pillars": [
    "state-affairs-environs",
    "newsbeat",
    "brands-smes",
    "opinions",
    "sports",
    "entertainment-lifestyle",
    "education",
    "thrills-tricks",
    "youth-corner",
    "tour-hotels"
  ],
  "corrections": {
    "BRAND & SMEs": "BRANDS & SMEs",
    "brand-smes": "brands-smes",
    "OPINION": "OPINIONS",
    "opinion": "opinions",
    "entertainment-lifestyle-world": "entertainment-lifestyle"
  },
  "direct_mappings": {
    "state-affairs-environs": "state-affairs-environs",
    "newsbeat": "newsbeat",
    "brand-smes": "brands-smes",
    "brands-smes": "brands-smes",
    "opinion": "opinions",
    "opinions": "opinions",
    "sports": "sports",
    "entertainment-lifestyle": "entertainment-lifestyle",
    "entertainment-lifestyle-world": "entertainment-lifestyle",
    "education": "education",
    "thrills-tricks": "thrills-tricks",
    "youth-corner": "youth-corner",
    "tour-hotels": "tour-hotels"
  },
  "non_pillar_fallbacks": {
    "art-comics": {
      "target_category": "entertainment-lifestyle",
      "add_tag": "Art & Comics"
    },
    "book-video-review": {
      "target_category": "entertainment-lifestyle",
      "add_tag": "Book & Video Review"
    },
    "events": {
      "target_category": "entertainment-lifestyle",
      "add_tag": "Events"
    },
    "fashion-fragrance": {
      "target_category": "entertainment-lifestyle",
      "add_tag": "Fashion & Fragrance"
    },
    "food-wine-2": {
      "target_category": "entertainment-lifestyle",
      "add_tag": "Food & Wine"
    },
    "healthy-living": {
      "target_category": "thrills-tricks",
      "add_tag": "Healthy Living"
    },
    "interviews": {
      "target_category": "opinions",
      "add_tag": "Interviews"
    },
    "personality-interview": {
      "target_category": "opinions",
      "add_tag": "Personality Interview"
    },
    "score-card-award": {
      "target_category": "state-affairs-environs",
      "add_tag": "Score Card & Award"
    },
    "uncategorized": {
      "target_category": "newsbeat",
      "add_tag": "Legacy Archive"
    }
  }
}
```
