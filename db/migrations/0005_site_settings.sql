-- Migration 0005: Site Copy and Broadcast Settings
-- Purpose: Persist configurable site copy (Hero, Ticker, Dossier, Network, Masthead) in the database

CREATE TABLE IF NOT EXISTS site_settings (
    key TEXT PRIMARY KEY,
    value JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default editorial broadcast settings
INSERT INTO site_settings (key, value)
VALUES (
    'editorial_copy',
    '{
        "hero": {
            "editionBadge": "CHERRYL MEDIA · DAILY EDITION",
            "titleLead": "Independent ",
            "titleAccent": "journalism",
            "titleTrail": " & digital media network.",
            "description": "Leading digital newsroom and new-age media distribution network covering Nigerian state affairs, business, culture, politics, and dispatches across Africa."
        },
        "ticker": {
            "label": "LIVE WIRE",
            "customNotice": "BREAKING: Federal Executive Council ratifies state electricity market guidelines and sovereign payment assurance frameworks."
        },
        "dossier": {
            "issueBadge": "WEEKEND DOSSIER · ISSUE Nº 18",
            "heading": "Culture, Cinema & Critical Essays.",
            "subheading": "Investigative photo essays, music production chronicles, and contemporary West African thought."
        },
        "network": {
            "headline": "Multi-Channel Media Distribution Network",
            "subheadline": "Reach over 250,000+ readers and professionals daily across West Africa and the diaspora through verified news channels.",
            "subscribers": "250,000+",
            "reachMonthly": "1.8M+"
        },
        "footer": {
            "tagline": "New-Age Media Distribution & Independent Journalism",
            "locations": "Abuja · Lagos · London",
            "contactEmail": "newsroom@cherrylmedia.com",
            "copyrightNotice": "© 2026 Cherryl Media Enterprise. All rights reserved."
        }
    }'::jsonb
)
ON CONFLICT (key) DO NOTHING;
