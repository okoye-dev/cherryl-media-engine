package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "postgresql://postgres.iwkduguiecqiicllskik:7Zryj7EfKrf7qHrO@aws-1-eu-central-1.pooler.supabase.com:6543/postgres"
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	// 1. Map articles to the 10 Canonical Pillars
	categoryUpdates := map[string]struct {
		Category     string
		CategoryName string
	}{
		"federal-executive-council-approves-unified-energy-framework": {
			Category:     "state-affairs-environs",
			CategoryName: "State Affairs & Environs",
		},
		"national-grid-stabilizes-with-5800mw-peak-feeder-band-overhaul": {
			Category:     "state-affairs-environs",
			CategoryName: "State Affairs & Environs",
		},
		"climate-resilience-act-subnational-adaptation-funds": {
			Category:     "state-affairs-environs",
			CategoryName: "State Affairs & Environs",
		},
		"sahelian-earth-hubs-passive-cooling": {
			Category:     "state-affairs-environs",
			CategoryName: "State Affairs & Environs",
		},
		"the-evolution-of-modular-solar-grids-in-northern-nigeria": {
			Category:     "state-affairs-environs",
			CategoryName: "State Affairs & Environs",
		},
		"federal-ministry-of-transport-finalizes-safety-standard": {
			Category:     "newsbeat",
			CategoryName: "Newsbeat",
		},
		"nigeria-electricity-regulatory-commission-updates-feeder-bands": {
			Category:     "newsbeat",
			CategoryName: "Newsbeat",
		},
		"distribution-companies-commence-phase-2-direct-meter-rollout": {
			Category:     "newsbeat",
			CategoryName: "Newsbeat",
		},
		"interstate-transit-digital-manifest-enforcement": {
			Category:     "newsbeat",
			CategoryName: "Newsbeat",
		},
		"cbn-leadership-announces-cross-border-fintech-settlement-rails": {
			Category:     "brands-smes",
			CategoryName: "Brands & SMEs",
		},
		"commercial-lenders-expand-sovereign-local-currency-trade-facilities": {
			Category:     "brands-smes",
			CategoryName: "Brands & SMEs",
		},
		"dual-listed-reits-unlock-commercial-property-liquidity": {
			Category:     "brands-smes",
			CategoryName: "Brands & SMEs",
		},
		"pencom-guidelines-allow-pension-equity-for-primary-mortgages": {
			Category:     "brands-smes",
			CategoryName: "Brands & SMEs",
		},
		"conversations-on-third-mainland-bridge-at-dusk": {
			Category:     "opinions",
			CategoryName: "Opinions",
		},
		"letters-on-awolowo-road-and-the-intimacy-of-modern-distance": {
			Category:     "opinions",
			CategoryName: "Opinions",
		},
		"midnight-calls-across-six-time-zones-diaspora-love": {
			Category:     "opinions",
			CategoryName: "Opinions",
		},
		"love-on-the-lekki-expressway-rush-hour-serendipity": {
			Category:     "opinions",
			CategoryName: "Opinions",
		},
		"inside-the-lagos-underground-polyphonic-afro-fusion": {
			Category:     "entertainment-lifestyle",
			CategoryName: "Entertainment & Lifestyle",
		},
		"art-x-lagos-finalizes-2026-pavilions": {
			Category:     "entertainment-lifestyle",
			CategoryName: "Entertainment & Lifestyle",
		},
		"lagos-soundscapes-and-creators-revolution": {
			Category:     "entertainment-lifestyle",
			CategoryName: "Entertainment & Lifestyle",
		},
		"preserving-highlife-wax-vintage-accra-sessions": {
			Category:     "entertainment-lifestyle",
			CategoryName: "Entertainment & Lifestyle",
		},
		"son-mandates-component-certification-for-rooftop-solar": {
			Category:     "thrills-tricks",
			CategoryName: "Thrills & Tricks",
		},
		"fermentation-coastal-smoke-heritage-indigenous-grains": {
			Category:     "tour-hotels",
			CategoryName: "Tour & Hotels",
		},
		"independent-comics-guild-rewriting-folklore": {
			Category:     "youth-corner",
			CategoryName: "Youth Corner",
		},
	}

	for slug, update := range categoryUpdates {
		res, err := pool.Exec(ctx, `
			UPDATE articles
			SET category = $1, updated_at = NOW()
			WHERE slug = $2
		`, update.Category, slug)
		if err != nil {
			log.Printf("error updating %s: %v", slug, err)
		} else {
			fmt.Printf("Updated %s -> %s (%d rows)\n", slug, update.Category, res.RowsAffected())
		}
	}

	// 2. Fetch UUIDs for section keys
	getIDs := func(slugs []string) []string {
		var ids []string
		for _, s := range slugs {
			var id string
			err := pool.QueryRow(ctx, "SELECT id::text FROM articles WHERE slug = $1", s).Scan(&id)
			if err == nil && id != "" {
				ids = append(ids, id)
			}
		}
		return ids
	}

	sectionsToSeed := map[string][]string{
		"hero_lead": {
			"national-grid-stabilizes-with-5800mw-peak-feeder-band-overhaul",
		},
		"hero_trending": {
			"cbn-leadership-announces-cross-border-fintech-settlement-rails",
			"federal-ministry-of-transport-finalizes-safety-standard",
			"commercial-lenders-expand-sovereign-local-currency-trade-facilities",
		},
		"state_affairs": {
			"federal-executive-council-approves-unified-energy-framework",
			"national-grid-stabilizes-with-5800mw-peak-feeder-band-overhaul",
			"climate-resilience-act-subnational-adaptation-funds",
			"sahelian-earth-hubs-passive-cooling",
		},
		"newsbeat": {
			"federal-ministry-of-transport-finalizes-safety-standard",
			"nigeria-electricity-regulatory-commission-updates-feeder-bands",
			"distribution-companies-commence-phase-2-direct-meter-rollout",
			"interstate-transit-digital-manifest-enforcement",
		},
		"brands_smes": {
			"cbn-leadership-announces-cross-border-fintech-settlement-rails",
			"commercial-lenders-expand-sovereign-local-currency-trade-facilities",
			"dual-listed-reits-unlock-commercial-property-liquidity",
			"pencom-guidelines-allow-pension-equity-for-primary-mortgages",
		},
		"opinions": {
			"conversations-on-third-mainland-bridge-at-dusk",
			"letters-on-awolowo-road-and-the-intimacy-of-modern-distance",
			"midnight-calls-across-six-time-zones-diaspora-love",
			"love-on-the-lekki-expressway-rush-hour-serendipity",
		},
		"weekend_dossier": {
			"lagos-soundscapes-and-creators-revolution",
			"art-x-lagos-finalizes-2026-pavilions",
			"inside-the-lagos-underground-polyphonic-afro-fusion",
			"fermentation-coastal-smoke-heritage-indigenous-grains",
		},
		"perspectives": {
			"conversations-on-third-mainland-bridge-at-dusk",
			"letters-on-awolowo-road-and-the-intimacy-of-modern-distance",
			"midnight-calls-across-six-time-zones-diaspora-love",
		},
		"dossier_lead": {
			"lagos-soundscapes-and-creators-revolution",
		},
		"dossier_soundscapes": {
			"inside-the-lagos-underground-polyphonic-afro-fusion",
		},
		"dossier_gastronomy": {
			"fermentation-coastal-smoke-heritage-indigenous-grains",
		},
		"dossier_radar": {
			"art-x-lagos-finalizes-2026-pavilions",
			"independent-comics-guild-rewriting-folklore",
			"preserving-highlife-wax-vintage-accra-sessions",
			"love-on-the-lekki-expressway-rush-hour-serendipity",
		},
	}

	for secKey, slugs := range sectionsToSeed {
		articleIDs := getIDs(slugs)
		_, err := pool.Exec(ctx, `
			INSERT INTO curated_sections (section_key, article_ids, metadata, updated_at)
			VALUES ($1, $2, '{}'::jsonb, NOW())
			ON CONFLICT (section_key) DO UPDATE
			SET article_ids = $2, updated_at = NOW()
		`, secKey, articleIDs)
		if err != nil {
			log.Printf("error saving section %s: %v", secKey, err)
		} else {
			fmt.Printf("Seeded section '%s' with %d articles\n", secKey, len(articleIDs))
		}
	}

	fmt.Println("\nAll canonical 10-pillar categories and curated sections updated successfully!")
}
