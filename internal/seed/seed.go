package seed

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/okoye-dev/cherryl-media-engine/internal/auth"
	"github.com/okoye-dev/cherryl-media-engine/internal/models"
)

type SeedableStore interface {
	Save(ctx context.Context, a *models.Article) error
	CreateUser(ctx context.Context, u *auth.User, passwordHash string) error
	SaveCandidate(ctx context.Context, s *models.SuggestedStory) error
}

type SectionSeeder interface {
	SaveSection(ctx context.Context, sectionKey string, articleIDs []string, metadata map[string]any) error
}

func Run(ctx context.Context, s SeedableStore) error {
	// 1. Seed Default Admin User
	adminPass := os.Getenv("INITIAL_ADMIN_PASSWORD")
	if adminPass == "" {
		adminPass = "cherry-" + uuid.New().String()[:8]
		log.Printf("[SEED] INITIAL_ADMIN_PASSWORD unset; generated temporary admin password: %s", adminPass)
	}
	adminPassHash, err := auth.HashPassword(adminPass)
	if err == nil {
		adminUser := &auth.User{
			ID:        "e0000000-0000-0000-0000-000000000001",
			Email:     "admin@cherrylmedia.com",
			FullName:  "Newsroom Editor",
			Role:      "admin",
			Active:    true,
			CreatedAt: time.Now().UTC(),
		}
		if err := s.CreateUser(ctx, adminUser, adminPassHash); err != nil {
			log.Printf("seed admin user: %v", err)
		} else {
			log.Println("seeded default admin user: admin@cherrylmedia.com")
		}
	}

	// 2. Seed All Admin Website Articles
	articles := GetAllWebsiteArticles()
	for _, art := range articles {
		a := art
		if err := s.Save(ctx, &a); err != nil {
			log.Printf("seed article %s: %v", a.Slug, err)
		}
	}
	log.Printf("seeded %d published articles", len(articles))

	// 3. Seed Suggested Discovery Leads
	candidates := GetInitialCandidates()
	for _, cand := range candidates {
		c := cand
		if err := s.SaveCandidate(ctx, &c); err != nil {
			log.Printf("seed candidate %s: %v", c.ID, err)
		}
	}
	log.Printf("seeded %d discovery leads", len(candidates))

	// 4. Seed Curated Sections
	if secSeeder, ok := s.(SectionSeeder); ok {
		defaultSections := []struct {
			Key string
			IDs []string
		}{
			{"hero_lead", []string{"2f405add-8bba-5261-99af-15feaa62754b"}},
			{"hero_trending", []string{
				"35588574-ec88-5a81-be90-8c46283c24de",
				"decd08ac-ef71-54ee-ac58-fd2480c007fa",
				"f10fad9c-c601-552e-a7b2-816b8c370ecc",
			}},
			{"featured_updates", []string{
				"25386df5-95e6-50ad-bc4c-3b6db589992d",
				"12f9e7fb-8019-5edb-80e9-7789a9729176",
				"d2065825-0408-5a7f-b4d0-559be9e67a9a",
				"278198f8-3eee-55ce-8736-dcdd55c46cec",
			}},
			{"perspectives", []string{
				"278198f8-3eee-55ce-8736-dcdd55c46cec",
				"d2065825-0408-5a7f-b4d0-559be9e67a9a",
				"f10fad9c-c601-552e-a7b2-816b8c370ecc",
			}},
			{"dossier_lead", []string{"f3033f02-9628-52d2-bef4-8b0814c6d835"}},
			{"dossier_soundscapes", []string{"d641aeae-a6be-50f5-a13c-54c67bbf6e75"}},
			{"dossier_gastronomy", []string{"0827a7d0-015b-5643-ac3d-4ccced2b933c"}},
			{"dossier_radar", []string{
				"f2fc6ae3-16ec-51cc-bc8d-c9d2b4e4271f",
				"ee3b36cb-8ce0-5f3c-bfae-3d077b882004",
				"917d0345-9020-50e5-89eb-9cb1721d21fa",
				"33a00b82-fd1f-5437-9317-0c951151d023",
			}},
		}
		for _, sec := range defaultSections {
			_ = secSeeder.SaveSection(ctx, sec.Key, sec.IDs, nil)
		}
		log.Printf("seeded %d curated sections", len(defaultSections))
	}

	return nil
}

func GetCanonicalArticles() []models.Article {
	now := time.Now().UTC()

	return []models.Article{
		{
			ID:           "art-01",
			Slug:         "federal-executive-council-approves-unified-energy-framework",
			Title:        "Federal Executive Council Approves Unified Energy and Distribution Guidelines",
			Kicker:       "NATIONAL POLICY ANNOUNCEMENT",
			Deck:         "Presidential council formalizes revised grid codes, state electricity market transitions, and sovereign payment assurance frameworks.",
			Category:     models.CategoryCurrentAffairs,
			CategoryName: "Nigerian Current Affairs",
			Author: models.Author{
				Name: "Bisi Alabi",
				Role: "Public Service & Utility Reporter",
			},
			FeaturedImage: models.FeaturedImage{
				URL:     "https://images.unsplash.com/photo-1541888946425-d0fbb186c5f7?auto=format&fit=crop&q=80&w=1200",
				Caption: "Federal Executive Council in session, Abuja.",
				Credit:  "State House Photo Archive",
			},
			ReadTime: "4 min read",
			Status:   models.StatusPublished,
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					Text:      "The Federal Executive Council (FEC) yesterday ratified the final operational framework for state-level electricity market devolution, unlocking statutory pathways for sub-national utility concessions.",
					IsDropCap: true,
				},
				{
					Type:  "heading",
					Level: 2,
					Text:  "Sovereign Payment Guarantees & Distribution Liquidity",
				},
				{
					Type: "paragraph",
					Text: "Under the gazetted directives, regional distribution companies will enter ring-fenced escrow agreements with state regulatory commissions to prevent liquidity shortfalls across transmission corridors.",
				},
				{
					Type:    "quote",
					Quote:   "This framework transitions Nigeria from a single central buyer bottleneck into competitive sub-national generation and retail markets.",
					Author:  "Ministerial Special Envoy on Power Infrastructure",
					Context: "Abuja State House Briefing",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "State House Official Press Gazette",
					URL:          "https://statehouse.gov.ng/releases/energy-framework-2026",
					PublishedAt:  "Oct 14, 2026",
					ClaimSummary: "Ratification of state-level electricity market guidelines and sovereign payment assurance.",
				},
				{
					SourceName:   "Federal Ministry of Power Briefing Notes",
					URL:          "https://power.gov.ng/briefings/grid-code-2026",
					PublishedAt:  "Oct 14, 2026",
					ClaimSummary: "Transmission loss mitigation and liquidity ring-fencing mechanism.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "FEC ratifies sovereign payment frameworks and decentralized grid codes for state electricity markets. Full policy analysis on Cherryl Media: https://cherrylmedia.com/article/federal-executive-council-approves-unified-energy-framework",
				Facebook:   "The Federal Executive Council has finalized the statutory guidelines allowing Nigerian states to independently license and regulate electricity markets. Read the complete verified dispatch.",
				LinkedIn:   "Institutional Briefing: Federal Executive Council formalizes sub-national energy market transitions with sovereign liquidity assurances.",
				Newsletter: "Our lead story this morning: Federal Executive Council formalizes state electricity markets and revised grid rules.",
			},
			PublishedAt: now.Add(-2 * time.Hour),
			CreatedAt:   now.Add(-2 * time.Hour),
		},
		{
			ID:           "art-02",
			Slug:         "nigeria-electricity-regulatory-commission-updates-feeder-bands",
			Title:        "NERC Releases Reclassified State Feeder Bands and Consumer Metering Schedule",
			Kicker:       "REGULATORY DISPATCH",
			Deck:         "Sub-station upgrades and state transmission telemetry dictate newly revised supply-hour minimums across regional distribution hubs.",
			Category:     models.CategoryPractical,
			CategoryName: "Practical Nigerian Updates",
			Author: models.Author{
				Name: "Bisi Alabi",
				Role: "Public Service & Utility Reporter",
			},
			FeaturedImage: models.FeaturedImage{
				URL:     "https://images.unsplash.com/photo-1509390144018-eeaf65052242?auto=format&fit=crop&q=80&w=1200",
				Caption: "High-voltage distribution transformer yard in Ikeja, Lagos.",
				Credit:  "Cherryl Media Staff File",
			},
			ReadTime: "3 min read",
			Status:   models.StatusPublished,
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					Text:      "The Nigerian Electricity Regulatory Commission (NERC) has published updated feeder band classifications, mandating real-time telemetry verification for Band A supply compliance across state capitals.",
					IsDropCap: true,
				},
				{
					Type: "paragraph",
					Text: "Customers on feeder corridors falling below the 20-hour benchmark for more than three consecutive cycles will automatically receive downward tariff adjustments credited against subsequent energy purchases.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "NERC Order on Feeder Alignment (No. 18/2026)",
					URL:          "https://nerc.gov.ng/orders/feeder-reclassification",
					PublishedAt:  "Oct 14, 2026",
					ClaimSummary: "Automated credit reimbursement for feeders failing 20-hour supply compliance.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "NERC publishes updated feeder classifications and introduces mandatory automated bill credits for supply shortfalls. Details: https://cherrylmedia.com/article/nigeria-electricity-regulatory-commission-updates-feeder-bands",
				Facebook:   "Check your feeder band: NERC mandates automatic billing credits for customers experiencing supply shortfalls.",
				LinkedIn:   "Regulatory Review: NERC enforces telemetry monitoring and automated reconciliation for regional utility customers.",
				Newsletter: "Practical Update: Verify your feeder band classification and compensation eligibility under new NERC guidelines.",
			},
			PublishedAt: now.Add(-5 * time.Hour),
			CreatedAt:   now.Add(-5 * time.Hour),
		},
		{
			ID:           "art-03",
			Slug:         "lagos-soundscapes-and-creators-revolution",
			Title:        "The Nollywood Studio Architecture: From Direct-to-Video Roots to Global Distribution Dominance",
			Kicker:       "CULTURE & ENTERPRISE",
			Deck:         "How independent Nigerian showrunners and production executives bypassed legacy syndication bottlenecks to build sovereign intellectual property empires.",
			Category:     models.CategoryEntertainment,
			CategoryName: "Entertainment & Internet Culture",
			Author: models.Author{
				Name: "Funke Balogun",
				Role: "Culture & Entertainment Bureau Chief",
			},
			FeaturedImage: models.FeaturedImage{
				URL:     "https://images.unsplash.com/photo-1536440136628-849c177e76a1?auto=format&fit=crop&q=80&w=1200",
				Caption: "Production rig on studio backlot, Lekki Phase 1, Lagos.",
				Credit:  "Lagos Soundstages Photo",
			},
			ReadTime: "7 min read",
			Status:   models.StatusPublished,
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					Text:      "For decades, the commercial reality of Nigerian cinema was defined by physical plastic cassettes traded in Idumota and Alaba International Market. Today, that guerrilla momentum has evolved into multi-million dollar studio campuses spanning Lagos and Ibadan.",
					IsDropCap: true,
				},
				{
					Type:    "quote",
					Quote:   "We stopped negotiating for shelf space. We engineered the audience first, and the global platforms had to come to us.",
					Author:  "Nollywood Executive Producer",
					Context: "Interview in Ikoyi, Lagos",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "National Film and Video Censors Board (NFVCB) Q3 Industry Report",
					URL:          "https://nfcb.gov.ng/reports/q3-2026",
					PublishedAt:  "Oct 12, 2026",
					ClaimSummary: "Domestic streaming licenses surpass international acquisitions for third consecutive quarter.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "From Alaba cassette stalls to sovereign intellectual property empires: The untold architecture of the new Nollywood studio. https://cherrylmedia.com/article/lagos-soundscapes-and-creators-revolution",
				Facebook:   "How Nigerian filmmakers engineered their own distribution pipelines to become Africa's premier storytelling powerhouse.",
				LinkedIn:   "Executive Analysis: Nollywood studio architecture, intellectual property retention, and streaming economics.",
				Newsletter: "Weekend Big Read: Inside the studio lots re-engineering West African cinema.",
			},
			PublishedAt: now.Add(-24 * time.Hour),
			CreatedAt:   now.Add(-24 * time.Hour),
		},
		{
			ID:           "art-04",
			Slug:         "central-bank-of-nigeria-fx-liquidity-guidelines",
			Title:        "Central Bank Refines Autonomous FX Market Rules for Enterprise Importers",
			Kicker:       "MONETARY POLICY",
			Deck:         "New circular establishes automated confirmation windows for verified letters of credit and manufacturing capital equipment.",
			Category:     models.CategoryBusiness,
			CategoryName: "Business & Institutions",
			Author: models.Author{
				Name: "Ibrahim Sani",
				Role: "Macroeconomics & Banking Editor",
			},
			FeaturedImage: models.FeaturedImage{
				URL:     "https://images.unsplash.com/photo-1611974789855-9c2a0a7236a3?auto=format&fit=crop&q=80&w=1200",
				Caption: "Central Bank of Nigeria Headquarters, Central Business District, Abuja.",
				Credit:  "CBN Communications",
			},
			ReadTime: "4 min read",
			Status:   models.StatusPublished,
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					Text:      "The Central Bank of Nigeria has issued directive TED/FEM/PUB/04/2026, creating streamlined electronic matching corridors for verified manufacturing inputs and industrial equipment.",
					IsDropCap: true,
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Central Bank of Nigeria Official Circular",
					URL:          "https://cbn.gov.ng/circulars/ted-fem-2026",
					PublishedAt:  "Oct 13, 2026",
					ClaimSummary: "Electronic matching corridors for capital equipment Letters of Credit.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "CBN circular creates fast-track electronic matching for manufacturing Letters of Credit. Analysis: https://cherrylmedia.com/article/central-bank-of-nigeria-fx-liquidity-guidelines",
				Facebook:   "Central Bank streamlines foreign exchange allocation for industrial inputs and raw materials.",
				LinkedIn:   "Macroeconomic Review: Central Bank updates autonomous window settlement protocols for enterprise importers.",
				Newsletter: "Business Desk: New CBN matching rules for enterprise trade finance.",
			},
			PublishedAt: now.Add(-36 * time.Hour),
			CreatedAt:   now.Add(-36 * time.Hour),
		},
		{
			ID:           "art-05",
			Slug:         "lagos-waterways-ferry-expansion-routes",
			Title:        "Lagos Waterways Authority Deploys New Electric Catamarans Across 6 Commuter Routes",
			Kicker:       "METROPOLITAN COMMUTE",
			Deck:         "Ikorodu to Marina transit times cut to 28 minutes as municipal ferry concessions add contactless multi-modal smart cards.",
			Category:     models.CategoryCurrentAffairs,
			CategoryName: "Nigerian Current Affairs",
			Author: models.Author{
				Name: "Bisi Alabi",
				Role: "Public Service & Utility Reporter",
			},
			FeaturedImage: models.FeaturedImage{
				URL:     "https://images.unsplash.com/photo-1544620347-c4fd4a3d5957?auto=format&fit=crop&q=80&w=1200",
				Caption: "Electric commuter ferry docking at Falomo Terminal, Ikoyi, Lagos.",
				Credit:  "LASWA Media",
			},
			ReadTime: "3 min read",
			Status:   models.StatusPublished,
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					Text:      "The Lagos State Waterways Authority (LASWA) has integrated six high-speed electric catamarans into the metropolitan transport network, providing emission-free transit across key lagoon corridors.",
					IsDropCap: true,
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Lagos State Ministry of Transportation Press Briefing",
					URL:          "https://transportation.lagosstate.gov.ng/laswa-ferry-launch",
					PublishedAt:  "Oct 11, 2026",
					ClaimSummary: "Deployment of 6 electric catamarans and Cowry card integration.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "LASWA deploys electric catamarans on Ikorodu-Marina routes with Cowry Card integration. Full route map: https://cherrylmedia.com/article/lagos-waterways-ferry-expansion-routes",
				Facebook:   "Commuter update: High-speed electric ferries now active between Ikorodu, Falomo, and Marina terminals.",
				LinkedIn:   "Urban Infrastructure: Lagos deploys electric passenger catamarans to expand metropolitan water transport.",
				Newsletter: "Commuter Dispatch: Faster lagoon routes with newly launched electric ferry services in Lagos.",
			},
			PublishedAt: now.Add(-48 * time.Hour),
			CreatedAt:   now.Add(-48 * time.Hour),
		},
	}
}

func GetInitialCandidates() []models.SuggestedStory {
	return []models.SuggestedStory{
		{
			ID:             uuid.New().String(),
			Headline:       "Sub-National Revenue Allocation: 12 States Post 40% Increase in Internally Generated Revenue",
			Pitch:          "State internal revenue services show significant gains following digital tax harmonisation and land title automation.",
			Category:       "business-institutions",
			RelevanceScore: 0.94,
			Sources: []models.SourceCitation{
				{
					SourceName:   "National Bureau of Statistics (NBS) State Fiscal Report",
					URL:          "https://nigerianstat.gov.ng/elibrary/read/1241",
					ClaimSummary: "12 states record greater than 40% YoY IGR growth driven by e-billing and property title titling.",
				},
			},
			ExtractedFacts: []string{
				"Ogun, Kaduna, and Edo states led the growth with digital property tax registries.",
				"Consolidated revenue accounts reduced collection leakages by an estimated 28%.",
			},
			Status:    "discovered",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:             uuid.New().String(),
			Headline:       "Universal Health Coverage: Primary Health Centres Modernisation Across South-East Corridors",
			Pitch:          "Revitalisation initiative combines solar cold-chain storage and telemedicine hubs across 34 rural local government areas.",
			Category:       "nigerian-current-affairs",
			RelevanceScore: 0.88,
			Sources: []models.SourceCitation{
				{
					SourceName:   "National Primary Health Care Development Agency (NPHCDA)",
					URL:          "https://nphcda.gov.ng/updates/south-east-revitalization",
					ClaimSummary: "Solar vaccine refrigerators installed in 120 rural clinic facilities.",
				},
			},
			ExtractedFacts: []string{
				"120 clinics equipped with solar-powered continuous temperature monitoring.",
				"Maternal health check-in times reduced by an average of 45 minutes.",
			},
			Status:    "discovered",
			CreatedAt: time.Now().Add(-3 * time.Hour),
		},
	}
}
