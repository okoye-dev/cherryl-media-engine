package testfixtures

import (
	"time"

	"github.com/google/uuid"
	"github.com/okoye-dev/cherryl-media-engine/internal/models"
)

// deterministicUUID generates a stable, reproducible UUID for an article slug.
func deterministicUUID(slug string) string {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("cherryl-article:"+slug)).String()
}

// GetAllWebsiteArticles compiles all canonical copy from the frontend formatted as complete backend Article models.
func GetAllWebsiteArticles() []models.Article {
	now := time.Now().UTC()
	_ = now

	return []models.Article{
		// 1. Federal Executive Council Confirms Unified National Clean Energy Transition Accord
		{
			ID:           deterministicUUID("federal-executive-council-approves-unified-energy-framework"),
			Slug:         "federal-executive-council-approves-unified-energy-framework",
			Title:        "Federal Executive Council Confirms Unified National Clean Energy Transition Accord",
			Kicker:       "POLICY DEVELOPMENT · GOVERNANCE",
			Deck:         "Confirmed policy shifts introduce immediate tariff restructuring, regional mini-grid guarantees, and multi-state transmission oversight protocols.",
			Category:     models.PillarCategory("nigerian-current-affairs"),
			CategoryName: "Current Affairs",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/e826f4da-b7b7-5120-bb2b-f466b33a929c.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/e826f4da-b7b7-5120-bb2b-f466b33a929c.webp",
				Caption:    "Ministerial council convenes during the extraordinary session in Abuja.",
				Credit:     "Federal Information Bureau",
			},
			ReadTime: "4 min read",
			Status:   models.StatusPublished,
			IsLead:   true,
			Tags:     []string{"FEC", "Energy", "Power", "Decentralization", "Policy", "Infrastructure"},
			Excerpt:  "The Federal Executive Council has ratified an inter-ministerial policy framework aimed at overhauling state-level power generation and establishing sub-national regulatory corridors.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "The Federal Executive Council has ratified an inter-ministerial policy framework aimed at overhauling state-level power generation and establishing sub-national regulatory corridors. Under the resolution finalized yesterday in Abuja, commercial developers will gain direct access to regional distribution networks without prior federal concession bottlenecks.",
				},
				{
					Type:    "quote",
					Quote:   "Our objective is to allow state governments and private capital to build dedicated industrial feeder lines without navigating redundant bureaucratic layers.",
					Author:  "Chief Adebayo Adelabu",
					Context: "Minister of Power",
				},
				{
					Type: "paragraph",
					Text: "Under the provisions of the 2023 Electricity Act now reaching full operational maturation, eight states have officially completed the transfer of regulatory oversight from the Nigerian Electricity Regulatory Commission (NERC) to state electricity regulatory boards. This decentralized posture is already triggering private liquidity inflows into localized gas-to-power and solar hybrid hubs across Ogun, Oyo, and Kano corridors.",
				},
				{
					Type:       "image",
					Src:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/ece0db31-b6f8-5eaf-9278-c901b46acfee.webp",
					StorageKey: "media/ece0db31-b6f8-5eaf-9278-c901b46acfee.webp",
					Caption:    "Regional mini-grid substation installation in Ogun State industrial corridor.",
					Credit:     "Photo: Cherryl Media Field Archive",
					Alt:        "Mini-grid solar distribution substation",
				},
				{
					Type:  "heading",
					Text:  "Fiscal Provisions and State Feeder Guarantees",
					Level: 2,
				},
				{
					Type: "paragraph",
					Text: "Crucially, the new framework mandates that distribution companies (DisCos) ring-fence industrial consumers on Band A feeders, ensuring that tariff pricing directly correlates with measured, metered uptime. Violations will trigger automated billing credits back to end consumers through unified metering software.",
				},
				{
					Type:     "internal_link",
					Title:    "NERC Releases Reclassified State Feeder Bands and Consumer Metering Schedule",
					Slug:     "nigeria-electricity-regulatory-commission-updates-feeder-bands",
					Category: "Practical Updates",
					Deck:     "Official public-service advisory detailing guaranteed supply hours, revised band thresholds, and direct customer dispute mechanisms.",
					ReadTime: "3 min read",
				},
				{
					Type: "paragraph",
					Text: "Economic observers note that while currency fluctuations continue to impact imported turbine equipment, long-term bilateral guarantees provided by sovereign infrastructure funds have mitigated immediate financing risk.",
				},
				{
					Type:      "external_link",
					Title:     "Federal Ministry of Power: Official Gazette on State Energy Decentralization",
					URL:       "https://nerc.gov.ng",
					Publisher: "NERC Regulatory Portal",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "State House Communications",
					URL:          "https://statehouse.gov.ng",
					PublishedAt:  "2026-10-14T08:00:00Z",
					ClaimSummary: "Ratification of inter-ministerial power generation devolution framework.",
				},
				{
					SourceName:   "NERC Official Gazette",
					URL:          "https://nerc.gov.ng",
					PublishedAt:  "2026-10-14T08:15:00Z",
					ClaimSummary: "Regulatory transfer to eight sub-national state electricity boards.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "BREAKING: Federal Executive Council ratifies unified national clean energy accord. State governments now empowered to licence commercial mini-grids directly. Here's what this means for industrial power tariffs 🧵👇",
				Facebook:   "The Federal Executive Council has officially confirmed a major shift in Nigeria's energy policy. State governments and private developers will now be able to build and regulate commercial mini-grids without prior federal delays. Read our full explainer on who is affected.",
				LinkedIn:   "Policy Analysis: Nigeria's Federal Executive Council has confirmed a unified clean energy transition framework, concluding the devolution of electricity regulation to sub-national commissions. For corporate energy buyers and infrastructure funds, this represents a structural change in off-grid power procurement.",
				Newsletter: "Today's Lead Story: The Federal Executive Council decentralized industrial power generation. We break down the tariff restructuring, state feeder guarantees, and investment implications.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 2. NERC Releases Reclassified State Feeder Bands and Consumer Metering Schedule
		{
			ID:           deterministicUUID("nigeria-electricity-regulatory-commission-updates-feeder-bands"),
			Slug:         "nigeria-electricity-regulatory-commission-updates-feeder-bands",
			Title:        "NERC Releases Reclassified State Feeder Bands and Consumer Metering Schedule",
			Kicker:       "CONSUMER ADVISORY · PUBLIC SERVICE",
			Deck:         "Official public-service advisory detailing guaranteed supply hours, revised band thresholds, and direct customer dispute mechanisms.",
			Category:     models.PillarCategory("practical-nigerian-updates"),
			CategoryName: "Practical Updates",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/f5aa1481-2d88-5159-9d90-91676df84733.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/f5aa1481-2d88-5159-9d90-91676df84733.webp",
				Caption:    "Technicians deploy smart meters in residential zones.",
				Credit:     "NERC Communications",
			},
			ReadTime: "3 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"NERC", "Electricity", "Metering", "Tariffs", "Consumer Rights", "Utilities"},
			Excerpt:  "The Nigerian Electricity Regulatory Commission has issued an updated national schedule for consumer feeder bands, clarifying minimum daily service thresholds across all eleven distribution franchises.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "The Nigerian Electricity Regulatory Commission has issued an updated national schedule for consumer feeder bands, clarifying minimum daily service thresholds and enforcement channels across all eleven electricity distribution franchises.",
				},
				{
					Type:  "heading",
					Text:  "Guaranteed Supply Hours by Feeder Classification",
					Level: 2,
				},
				{
					Type: "paragraph",
					Text: "Under the reclassification: Band A feeders guarantee a minimum of 20 hours daily; Band B requires 16 hours; Band C requires 12 hours. Any feeder failing to meet its guaranteed threshold for three consecutive days is subject to automatic downgrade and retrospective billing adjustments.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "NERC Consumer Affairs Division",
					URL:          "https://nerc.gov.ng",
					PublishedAt:  "2026-10-14T09:00:00Z",
					ClaimSummary: "Feeder band schedule update and consumer compensation protocol.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "ELECTRICITY UPDATE: NERC has released the updated feeder band schedules. Band A (20+ hrs), Band B (16 hrs), Band C (12 hrs). What to do if your DisCo falls short of guaranteed hours ⚡️👇",
				Facebook:   "Are you paying Band A electricity tariffs but getting less than 20 hours daily? NERC's revised regulatory schedule introduces mandatory downgrades and credit refunds. Read our step-by-step consumer guide.",
				LinkedIn:   "Utility Regulation: NERC issues reclassified consumer feeder schedules with automated billing refund requirements for DisCos. Operational summary for facility managers and manufacturing facilities.",
				Newsletter: "Consumer Advisory: NERC's reclassified feeder schedules are live. Here are the guaranteed hours, tariff bands, and the direct complaint escalation link.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 3. The Nollywood Studio Architecture: From Direct-to-Video Roots to Global Distribution Dominance
		{
			ID:           deterministicUUID("lagos-soundscapes-and-creators-revolution"),
			Slug:         "lagos-soundscapes-and-creators-revolution",
			Title:        "The Nollywood Studio Architecture: From Direct-to-Video Roots to Global Distribution Dominance",
			Kicker:       "WEEKEND DOSSIER · CULTURE",
			Deck:         "How independent Nigerian showrunners, cinematographers, and production executives bypassed legacy syndication bottlenecks to build sovereign intellectual property empires.",
			Category:     models.PillarCategory("entertainment-internet-culture"),
			CategoryName: "Entertainment & Internet Culture",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/fff4d6db-84bb-523d-ae18-2d8451490929.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/fff4d6db-84bb-523d-ae18-2d8451490929.webp",
				Caption:    "High-definition soundstage in a multi-stage production complex in Epe.",
				Credit:     "Lagos Soundstages Photo Archive",
			},
			ReadTime: "7 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Nollywood", "Film", "Culture", "Cinema", "Lagos", "Intellectual Property"},
			Excerpt:  "How independent Nigerian showrunners, cinematographers, and production executives bypassed legacy syndication bottlenecks to build sovereign intellectual property empires.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "In a cavernous soundstage off the Lekki-Epe corridor, a seventy-person crew works under silent LED rigs. Thirty years ago, their predecessors wrapped feature films in forty-eight hours on Hi8 videotape sold out of roadside stalls in Idumota.",
				},
				{
					Type:    "quote",
					Quote:   "We stopped negotiating for shelf space. We engineered the audience first, and then the global streamers had no choice but to come to Lagos.",
					Author:  "Kemi Adesanya",
					Context: "Showrunner & Producer",
				},
				{
					Type:       "image",
					Src:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/3d6b3a23-ffaf-5df6-82db-f51fa5077518.webp",
					StorageKey: "media/3d6b3a23-ffaf-5df6-82db-f51fa5077518.webp",
					Caption:    "Audience premiere at a contemporary West African cinema festival.",
					Credit:     "Photo: Film Guild Archives",
					Alt:        "Cinema festival audience premiere",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Nigerian Film Corporation Archive",
					URL:          "https://filmnigeria.gov.ng",
					PublishedAt:  "2026-10-12T14:00:00Z",
					ClaimSummary: "Annual theatrical admissions and digital distribution data.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "ESSAY: From Idumota cassette hawkers to global streaming acquisitions — how Nollywood directors engineered a sovereign intellectual property empire 🎬✨",
				Facebook:   "Thirty years after the first VHS blockbusters were sold out of market stalls in Lagos, Nigerian cinema is dominating global streaming charts. Funke Balogun reports on the studio architecture reshaping West African storytelling.",
				LinkedIn:   "Creative Economy Analysis: Nollywood's structural evolution from direct-to-video distribution to multi-million dollar co-production frameworks. An inside look at production finance across West Africa.",
				Newsletter: "Weekend Dossier: The architecture of contemporary Nollywood. How Nigerian filmmakers built a billion-dollar creative export engine from scratch.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 4. CBN Finalizes Real-Time Cross-Border Settlement Framework with Regional Central Banks
		{
			ID:           deterministicUUID("cbn-leadership-announces-cross-border-fintech-settlement-rails"),
			Slug:         "cbn-leadership-announces-cross-border-fintech-settlement-rails",
			Title:        "CBN Finalizes Real-Time Cross-Border Settlement Framework with Regional Central Banks",
			Kicker:       "FINANCIAL SYSTEM · MARKETS",
			Deck:         "Central Bank of Nigeria collaborates with West African Monetary Agency to enable direct local-currency wholesale clearing across eight national jurisdictions.",
			Category:     models.PillarCategory("business-institutions"),
			CategoryName: "Business & Institutions",
			Author: models.Author{
				Name: "Chukwudi Nwosu",
				Role: "Capital Markets Analyst",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/a2cf6f5d-8512-576f-b87f-8364d19d46ee.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/a2cf6f5d-8512-576f-b87f-8364d19d46ee.webp",
				Caption:    "Regional central bank governors sign multilateral settlement pact in Abuja.",
				Credit:     "CBN Corporate Communications",
			},
			ReadTime: "5 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"CBN", "Banking", "Fintech", "PAPSS", "Foreign Exchange", "Trade"},
			Excerpt:  "The Central Bank of Nigeria has finalized technical protocols linking domestic instant payment switches with regional settlement corridors across West Africa.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "The Central Bank of Nigeria has concluded multilateral integration protocols linking domestic instant payment switches with the Pan-African Payment and Settlement System (PAPSS).",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Central Bank of Nigeria",
					URL:          "https://cbn.gov.ng",
					PublishedAt:  "2026-10-14T11:00:00Z",
					ClaimSummary: "Multilateral settlement protocol ratification.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "FINTECH & MARKETS: CBN connects Nigerian banking switches to regional clearing rails. Merchants can now settle West African cross-border trade directly in Naira 💳📊",
				Facebook:   "Direct regional trade settlement without third-party correspondent banks: CBN finalizes integration with PAPSS. Here's what commercial traders need to know.",
				LinkedIn:   "Monetary Policy & Trade: Central Bank of Nigeria concludes real-time cross-border settlement rails. Strategic analysis on foreign exchange reserves and intra-African liquidity.",
				Newsletter: "Financial Markets Wire: The Central Bank of Nigeria just cleared the path for direct Naira-denominated trade across West Africa.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 5. Conversations on Third Mainland Bridge at Dusk: Reflections on Modern Longing in Lagos
		{
			ID:           deterministicUUID("conversations-on-third-mainland-bridge-at-dusk"),
			Slug:         "conversations-on-third-mainland-bridge-at-dusk",
			Title:        "Conversations on Third Mainland Bridge at Dusk: Reflections on Modern Longing in Lagos",
			Kicker:       "STORIES & ROMANCE · ESSAYS",
			Deck:         "Between the humid salt air of the lagoon and taillights stretching toward Oworonshoki, two travelers dissect what it means to stay tethered to someone across twenty million people.",
			Category:     models.PillarCategory("stories-romance"),
			CategoryName: "Stories & Romance",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/99f72c5d-6c90-5d7e-a61d-2928952e13e7.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/99f72c5d-6c90-5d7e-a61d-2928952e13e7.webp",
				Caption:    "Amber dusk settling over the Lagos lagoon and Third Mainland Bridge.",
				Credit:     "Lagos Lens Photographic Archive",
			},
			ReadTime: "6 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Memoir", "Lagos", "Romance", "Essay", "Narrative", "City"},
			Excerpt:  "At 6:45 PM on a Tuesday, Third Mainland Bridge ceases to be a concrete thoroughfare and becomes a collective holding space for eight million unresolved thoughts.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "At 6:45 PM on a Tuesday, Third Mainland Bridge ceases to be an eleven-kilometer concrete artery and becomes a collective holding chamber for eight million private longings.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Cherryl Literary Desk",
					URL:          "https://cherrylmedia.com",
					PublishedAt:  "2026-10-13T18:00:00Z",
					ClaimSummary: "Personal narrative and cultural memoir.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "ESSAY: 'In Lagos, loving someone across twenty million people feels less like romance and more like an act of defiance.' Conversations on Third Mainland Bridge at dusk 🌆🌊",
				Facebook:   "Between the humid salt air of the lagoon and taillights stretching toward Oworonshoki, Amara Okonkwo writes on courtship, distance, and memory in modern Lagos.",
				LinkedIn:   "Cultural Perspectives: Urban geography and human intimacy in megacities. A narrative essay on life along the Lagos lagoon.",
				Newsletter: "Evening Essay: Conversations on Third Mainland Bridge at Dusk. An evocative exploration of courtship and modern distance.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 6. Federal Ministry of Transport Directs Regional Rail Link Feasibility Review
		{
			ID:           deterministicUUID("federal-ministry-of-transport-finalizes-safety-standard"),
			Slug:         "federal-ministry-of-transport-finalizes-safety-standard",
			Title:        "Federal Ministry of Transport Directs Regional Rail Link Feasibility Review",
			Kicker:       "PUBLIC SERVICE · INFRASTRUCTURE",
			Deck:         "Technical directives mandate passenger coach telemetry and unified regional corridor signaling systems across southern and middle-belt transit networks.",
			Category:     models.PillarCategory("nigerian-current-affairs"),
			CategoryName: "Current Affairs",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/bc51d838-573d-5947-9372-ac03075b98e8.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/bc51d838-573d-5947-9372-ac03075b98e8.webp",
				Caption:    "Locomotive engineers inspect passenger coaches at the Ebute Metta terminal.",
				Credit:     "Federal Ministry of Transport",
			},
			ReadTime: "3 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Rail", "Transport", "Infrastructure", "Safety", "Federal Government"},
			Excerpt:  "The Federal Ministry of Transport has ordered an immediate technical assessment of passenger safety systems across operational rail routes.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "The Federal Ministry of Transport has ordered an immediate technical assessment of signaling infrastructure across standard-gauge passenger corridors.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Federal Ministry of Transport",
					URL:          "https://transportation.gov.ng",
					PublishedAt:  "2026-10-14T07:30:00Z",
					ClaimSummary: "Rail telemetry directives release.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "RAIL SAFETY: Ministry of Transport directs comprehensive telemetry and signaling review across national rail lines 🚆📋",
				Facebook:   "The Federal Ministry of Transport has issued new mandatory guidelines for regional passenger rail operations. Read full safety schedule.",
				LinkedIn:   "Infrastructure Watch: Federal Ministry of Transport directives mandate passenger rail signaling overhaul.",
				Newsletter: "Transit Update: New federal directives introduce unified telemetry standards for passenger rail.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 7. DisCos Issue Consumer Advisory on Prepaid Smart Meter Firmware Upgrades
		{
			ID:           deterministicUUID("distribution-companies-commence-phase-2-direct-meter-rollout"),
			Slug:         "distribution-companies-commence-phase-2-direct-meter-rollout",
			Title:        "DisCos Issue Consumer Advisory on Prepaid Smart Meter Firmware Upgrades",
			Kicker:       "PRACTICAL ADVISORY",
			Deck:         "Ten million prepaid meters to receive automatic over-the-air firmware upgrades; distribution utilities publish schedule by state and ward.",
			Category:     models.PillarCategory("practical-nigerian-updates"),
			CategoryName: "Practical Updates",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/4958ea81-ae31-51f0-a893-74090ea4b453.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/4958ea81-ae31-51f0-a893-74090ea4b453.webp",
				Caption:    "Utility engineers configure prepaid smart meter communication modules.",
				Credit:     "DisCo Operations Consortium",
			},
			ReadTime: "4 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Meters", "DisCos", "Energy", "Consumer", "SmartGrid"},
			Excerpt:  "Distribution companies have published the phase 2 execution plan for prepaid meter TID rollover and firmware upgrades across all 36 states.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "Electricity distribution companies across Nigeria have published detailed schedules for the mandatory smart meter firmware migration.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Association of Nigerian Electricity Distributors",
					URL:          "https://anedng.com",
					PublishedAt:  "2026-10-14T10:00:00Z",
					ClaimSummary: "Phase 2 meter upgrade release.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "CONSUMER ADVISORY: DisCos publish phase 2 smart meter rollover schedule. Check if your state is affected this week 💡",
				Facebook:   "Prepaid meter updates: The distribution companies have announced the timeline for the phase 2 smart meter upgrade rollout.",
				LinkedIn:   "Utility Operations: DisCos announce phase 2 prepaid meter rollout and firmware compliance schedule.",
				Newsletter: "Consumer Alert: Smart meter firmware upgrade schedules released by distribution utilities.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 8. Commercial Lenders Expand Sovereign Local-Currency Trade Facilities
		{
			ID:           deterministicUUID("commercial-lenders-expand-sovereign-local-currency-trade-facilities"),
			Slug:         "commercial-lenders-expand-sovereign-local-currency-trade-facilities",
			Title:        "Commercial Lenders Expand Sovereign Local-Currency Trade Facilities",
			Kicker:       "CAPITAL MARKETS",
			Deck:         "Commercial banking syndicate structures N500bn trade financing liquidity pool for manufacturing exporters and agro-allied processors.",
			Category:     models.PillarCategory("business-institutions"),
			CategoryName: "Business & Institutions",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/d44aee35-6b1f-5212-bc28-f78f1f061864.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/d44aee35-6b1f-5212-bc28-f78f1f061864.webp",
				Caption:    "Commercial banking trading desk in Victoria Island, Lagos.",
				Credit:     "Financial Markets Archive",
			},
			ReadTime: "5 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Banking", "Finance", "Exports", "Lenders", "Liquidity", "Business"},
			Excerpt:  "A syndicate of leading Nigerian tier-one financial institutions has created a dedicated domestic trade credit facility to support non-oil exporters.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "A consortium of commercial lenders has formalized a 500-billion-naira credit facility targeted at non-oil exporters and local industrial fabricators.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Bankers Committee Statement",
					URL:          "https://cbn.gov.ng",
					PublishedAt:  "2026-10-14T12:00:00Z",
					ClaimSummary: "Local currency export facility launch.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "MARKETS: Commercial banks announce N500bn credit facility for non-oil exporters and agro-processors 📈💵",
				Facebook:   "N500 billion liquidity pool launched: Nigerian lenders announce single-digit credit windows for local manufacturing exporters.",
				LinkedIn:   "Capital Markets: Tier-1 banking syndicate establishes N500bn non-oil export financing pool.",
				Newsletter: "Business Desk: N500bn trade facility launched for manufacturing and agro-allied processors.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 9. Letters on Awolowo Road: The Intimacy of Modern Distance in Ikoyi
		{
			ID:           deterministicUUID("letters-on-awolowo-road-and-the-intimacy-of-modern-distance"),
			Slug:         "letters-on-awolowo-road-and-the-intimacy-of-modern-distance",
			Title:        "Letters on Awolowo Road: The Intimacy of Modern Distance in Ikoyi",
			Kicker:       "STORIES & ROMANCE",
			Deck:         "An evocative personal reflection on courtship, late-night dinners, and navigating emotional distance in Lagos.",
			Category:     models.PillarCategory("stories-romance"),
			CategoryName: "Stories & Romance",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/21b50b9b-85ec-5dae-9116-ebade33a41e8.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/21b50b9b-85ec-5dae-9116-ebade33a41e8.webp",
				Caption:    "Rain glistening on the asphalt outside a cafe on Awolowo Road, Ikoyi.",
				Credit:     "Urban Lens Archive",
			},
			ReadTime: "4 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Romance", "Ikoyi", "Lagos", "Memoir", "Storytelling", "Culture"},
			Excerpt:  "We spent three hours talking over iced coffee at a corner table on Awolowo Road, pretending the rain outside was the only reason we hadn't called a cab.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "We spent three hours talking over iced coffee at a corner table on Awolowo Road, pretending the heavy October downpour outside was the only reason we hadn't called a ride home.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Cherryl Literary Desk",
					URL:          "https://cherrylmedia.com",
					PublishedAt:  "2026-10-13T16:00:00Z",
					ClaimSummary: "Literary essay on urban connection.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "ESSAY: Late-afternoon rain on Awolowo Road, cold brews, and the fragile geometry of falling in love in Lagos 🌧️☕",
				Facebook:   "Read Amara Okonkwo's latest essay on courtship and quiet intimacy on the tree-lined streets of Ikoyi.",
				LinkedIn:   "Cultural Feature: Reflections on modern connection and urban life in Lagos.",
				Newsletter: "Stories & Romance: Letters on Awolowo Road. A quiet reflection on intimacy in Ikoyi.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 10. Inside the Lagos Underground: The Sound Architects Engineering Polyphonic Afro-Fusion
		{
			ID:           deterministicUUID("inside-the-lagos-underground-polyphonic-afro-fusion"),
			Slug:         "inside-the-lagos-underground-polyphonic-afro-fusion",
			Title:        "Inside the Lagos Underground: The Sound Architects Engineering Polyphonic Afro-Fusion",
			Kicker:       "SOUNDSCAPES & ACOUSTICS",
			Deck:         "A technical exploration into the rhythmic syncopations and modular synthesizers defining new West African acoustics.",
			Category:     models.PillarCategory("entertainment-internet-culture"),
			CategoryName: "Entertainment & Internet Culture",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/68d885ad-623f-5769-b9ea-6deb64c1b762.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/68d885ad-623f-5769-b9ea-6deb64c1b762.webp",
				Caption:    "Analog synthesizer patchbay in a recording sanctuary in Surulere.",
				Credit:     "Lagos Soundstages Photo",
			},
			ReadTime: "4 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"AfroFusion", "MusicProduction", "Soundscapes", "Synthesizers", "Lagos"},
			Excerpt:  "Behind heavy soundproof doors in Surulere, a new wave of sound architects is blending traditional Yoruba talking drums with modular analog synthesizers.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "Behind heavy soundproof doors in Surulere and Ebute Metta, a new wave of electronic composers is blending centuries-old polyrhythms with vintage Japanese analog synthesizers.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Lagos Sound Archive",
					URL:          "https://cherrylmedia.com",
					PublishedAt:  "2026-10-13T11:00:00Z",
					ClaimSummary: "Acoustic production methodology.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "MUSIC TECH: Talking drums meet modular analog synthesizers. Inside the Surulere studios defining polyphonic Afro-Fusion 🎧🎛️",
				Facebook:   "Funke Balogun takes readers inside the underground music laboratories in Surulere reshaping contemporary West African audio engineering.",
				LinkedIn:   "Audio Engineering & Media: Technical study of hybrid analog synthesis in contemporary African music production.",
				Newsletter: "Culture Dossier: The sound architects building the future of Afro-Fusion.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 11. Fermentation, Coastal Smoke & Heritage: The Contemporary Revival of Indigenous Grains
		{
			ID:           deterministicUUID("fermentation-coastal-smoke-heritage-indigenous-grains"),
			Slug:         "fermentation-coastal-smoke-heritage-indigenous-grains",
			Title:        "Fermentation, Coastal Smoke & Heritage: The Contemporary Revival of Indigenous Grains",
			Kicker:       "ENTERTAINMENT & LIFESTYLE",
			Deck:         "Culinary anthropologists and contemporary chefs recharting regional biodiversity across coastal markets and ancestral fermentation pits.",
			Category:     models.PillarCategory("entertainment-internet-culture"),
			CategoryName: "Entertainment & Internet Culture",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/a3bd1c2d-ca6a-5a2c-900d-75461a3a4078.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/a3bd1c2d-ca6a-5a2c-900d-75461a3a4078.webp",
				Caption:    "Heirloom sorghum and fonio grains curated in an artisanal culinary research kitchen.",
				Credit:     "West African Botanical Archive",
			},
			ReadTime: "5 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Gastronomy", "Heritage", "Grains", "Culinary", "Fermentation", "Food"},
			Excerpt:  "Culinary anthropologists and contemporary chefs recharting regional biodiversity across coastal markets.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "Across test kitchens in Victoria Island and traditional earth kitchens along the Badagry coastline, a quiet culinary revolution is reclaiming fonio, heirloom millet, and fermented locust beans from culinary marginality.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Culinary Heritage Institute",
					URL:          "https://cherrylmedia.com",
					PublishedAt:  "2026-10-13T13:00:00Z",
					ClaimSummary: "Documentation of indigenous crop agronomy.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "GASTRONOMY: Fonio, heirloom millet, and ancestral fermentation: How West African chefs are reclaiming native biodiversity 🌾🍲",
				Facebook:   "From coastal smokehouses to tasting menus in Lagos: Funke Balogun explores the indigenous grain movement reviving historic African agricultural heritage.",
				LinkedIn:   "Agro-Culinary Innovation: The revival of drought-resilient indigenous grains across West Africa.",
				Newsletter: "Culture Wire: Heritage grains, coastal smoke, and the rebirth of indigenous gastronomy.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 12. Art X Lagos Finalizes 2026 Pavilions with 42 International Galleries
		{
			ID:           deterministicUUID("art-x-lagos-finalizes-2026-pavilions"),
			Slug:         "art-x-lagos-finalizes-2026-pavilions",
			Title:        "Art X Lagos Finalizes 2026 Pavilions with 42 International Galleries",
			Kicker:       "VISUAL ARTS · EXHIBITIONS",
			Deck:         "West Africa's premier international art fair expands institutional programming, featuring pioneering digital sculpture and textile preservation pavilions.",
			Category:     models.PillarCategory("entertainment-internet-culture"),
			CategoryName: "Entertainment & Internet Culture",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/a439611c-2737-5bf1-8b0b-c5933ceee3e3.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/a439611c-2737-5bf1-8b0b-c5933ceee3e3.webp",
				Caption:    "Curators install large-format sculptural works at the Federal Palace pavilion.",
				Credit:     "Art X Lagos Communications",
			},
			ReadTime: "2 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Art", "ArtXLagos", "Exhibitions", "Galleries", "Culture"},
			Excerpt:  "West Africa's flagship art fair confirms the participation of leading contemporary spaces across Africa and the global diaspora.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "Art X Lagos has finalized its curated exhibitor roster for the upcoming 2026 edition, welcoming forty-two galleries from twenty-six nations.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Art X Lagos Official Press Room",
					URL:          "https://artxlagos.com",
					PublishedAt:  "2026-10-14T14:00:00Z",
					ClaimSummary: "Pavilion announcement and curated gallery list.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "ART RADAR: 42 international galleries, monumental sculptures, and textile innovations. Art X Lagos confirms 2026 edition 🎨✨",
				Facebook:   "Art X Lagos returns with 42 international galleries representing contemporary creators from 26 countries.",
				LinkedIn:   "Cultural Economy: Art X Lagos finalizes 2026 exhibitor list, reflecting institutional liquidity in African contemporary art.",
				Newsletter: "Culture Brief: Art X Lagos finalizes 2026 pavilions.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 13. Preserving Highlife Wax: The Vinyl Restorers Digitizing Vintage Accra Sessions
		{
			ID:           deterministicUUID("preserving-highlife-wax-vintage-accra-sessions"),
			Slug:         "preserving-highlife-wax-vintage-accra-sessions",
			Title:        "Preserving Highlife Wax: The Vinyl Restorers Digitizing Vintage Accra Sessions",
			Kicker:       "ARCHIVES & SOUND PRESERVATION",
			Deck:         "Sound archivists and audio restorers rescue irreplaceable 1960s master acetates and wax pressings from tropical heat and degradation.",
			Category:     models.PillarCategory("entertainment-internet-culture"),
			CategoryName: "Entertainment & Internet Culture",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/c378c8dc-0c2f-5281-ae2e-8f920bde290d.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/c378c8dc-0c2f-5281-ae2e-8f920bde290d.webp",
				Caption:    "Turntable optical scanner reading microgrooves on a rare highlife 45rpm record.",
				Credit:     "Sound Preservation Trust",
			},
			ReadTime: "4 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Highlife", "Vinyl", "AudioArchiving", "Accra", "MusicHistory"},
			Excerpt:  "Audio restoration engineers in Accra and Lagos are racing to convert fragile shellac and vinyl recordings into lossless digital masters.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "In an air-conditioned laboratory in central Accra, audio engineer Kofi Mensah uses optical laser turntables to clean and digitize fragile 1960s highlife records.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "West African Sound Preservation Trust",
					URL:          "https://cherrylmedia.com",
					PublishedAt:  "2026-10-13T09:00:00Z",
					ClaimSummary: "Archive preservation registry data.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "SOUND ARCHIVES: Optical lasers, ultrasonic baths, and 1960s Accra masters: The restorers saving Highlife vinyl before it decays 🎵📻",
				Facebook:   "Inside the laboratory racing against tropical humidity to preserve the golden age of West African highlife vinyl recordings.",
				LinkedIn:   "Cultural Heritage: Preservation of historic musical archives through advanced optical audio restoration.",
				Newsletter: "Archive Dispatch: Saving the vintage highlife sessions of Accra and Lagos.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 14. Sahelian Earth Hubs: Passive Cooling Innovations in Urban Architecture
		{
			ID:           deterministicUUID("sahelian-earth-hubs-passive-cooling"),
			Slug:         "sahelian-earth-hubs-passive-cooling",
			Title:        "Sahelian Earth Hubs: Passive Cooling Innovations in Urban Architecture",
			Kicker:       "ARCHITECTURE & SUSTAINABILITY",
			Deck:         "Contemporary architects fuse compressed earth blocks, wind towers, and bioclimatic design to drop indoor temperatures without air conditioning.",
			Category:     models.PillarCategory("nigerian-current-affairs"),
			CategoryName: "Current Affairs",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/fa264975-1c61-5ce2-8f5b-ce07127defe5.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/fa264975-1c61-5ce2-8f5b-ce07127defe5.webp",
				Caption:    "Bioclimatic earth-block community hub in northern Nigeria featuring natural airflow chimneys.",
				Credit:     "Sustainable Sahel Architecture Bureau",
			},
			ReadTime: "3 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Architecture", "Sustainability", "Sahel", "Cooling", "Urbanism"},
			Excerpt:  "New civic and community facilities built with compressed earth masonry achieve 8-degree Celsius indoor temperature reductions naturally.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "Architects across Kaduna, Kano, and Niamey are demonstrating how centuries-old Sahelian earthen masonry and thermal mass principles can solve modern energy crises.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Nigerian Institute of Architects Technical Bulletin",
					URL:          "https://nia.ng",
					PublishedAt:  "2026-10-12T16:00:00Z",
					ClaimSummary: "Bioclimatic thermal performance study.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "SPATIAL DESIGN: Sahelian wind towers and compressed earth blocks: dropping indoor temperatures by 8°C without relying on the grid 🏛️🌬️",
				Facebook:   "Can ancient earthen masonry solve modern West African cooling challenges? How architects are innovating with bioclimatic earth construction.",
				LinkedIn:   "Sustainable Infrastructure: Sahelian bioclimatic architecture and energy-neutral urban design.",
				Newsletter: "Architecture Dispatch: Passive cooling innovations across Sahelian urban hubs.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 15. Independent Comics Guild Rewriting Regional Folklore for Young Audiences
		{
			ID:           deterministicUUID("independent-comics-guild-rewriting-folklore"),
			Slug:         "independent-comics-guild-rewriting-folklore",
			Title:        "Independent Comics Guild Rewriting Regional Folklore for Young Audiences",
			Kicker:       "GRAPHIC NOVELS · PUBLISHING",
			Deck:         "Young illustrators and writers across Lagos and Ibadan bypass legacy publishing gatekeepers to print serialized comic anthologies rooted in local lore.",
			Category:     models.PillarCategory("entertainment-internet-culture"),
			CategoryName: "Entertainment & Internet Culture",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/64185597-9d0e-55ba-b4fc-d997c55868ad.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/64185597-9d0e-55ba-b4fc-d997c55868ad.webp",
				Caption:    "Original ink sketches and character designs from the West African Comic Anthology.",
				Credit:     "Independent Comics Guild",
			},
			ReadTime: "3 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Comics", "Folklore", "Illustration", "Publishing", "Culture"},
			Excerpt:  "Independent graphic novelists are crafting richly illustrated serialized adventures steeped in Yoruba, Igbo, and Hausa mythology.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "In print shops across Somolu, young graphic novelists and colorists are producing limited-edition serialized comic books that reimagine ancient folklore for digital natives.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "African Comic Creators Collective",
					URL:          "https://cherrylmedia.com",
					PublishedAt:  "2026-10-14T15:00:00Z",
					ClaimSummary: "Independent publishing catalog.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "GRAPHIC NOVELS: Somolu print shops, vibrant character design, and indigenous mythology. The independent comics wave taking over Lagos 📚💥",
				Facebook:   "Meet the young illustrators and graphic novelists rewriting regional folklore for a whole new generation of readers.",
				LinkedIn:   "Publishing & Media: The growth of independent comic serialization and grassroots publishing across Nigeria.",
				Newsletter: "Creative Dispatch: Independent comics guild rewrites regional folklore.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 16. National Climate Council Disburses First Tranche of Subnational Adaptation Grants
		{
			ID:           deterministicUUID("climate-resilience-act-subnational-adaptation-funds"),
			Slug:         "climate-resilience-act-subnational-adaptation-funds",
			Title:        "National Climate Council Disburses First Tranche of Subnational Adaptation Grants",
			Kicker:       "CLIMATE GOVERNANCE · PUBLIC POLICY",
			Deck:         "Sixteen ecological disaster-vulnerable states receive targeted matching capital for flood plain defense, mangrove restoration, and municipal drainage infrastructure.",
			Category:     models.PillarCategory("nigerian-current-affairs"),
			CategoryName: "Current Affairs",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/1b8e4f20-9111-5333-8222-777788889999.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/1b8e4f20-9111-5333-8222-777788889999.webp",
				Caption:    "Ecological remediation and flood barrier construction works across frontline subnational zones.",
				Credit:     "National Climate Secretariat",
			},
			ReadTime: "4 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Climate", "Policy", "Infrastructure", "Ecology", "Governance"},
			Excerpt:  "The National Climate Council has finalized disbursement modalities for subnational ecological defense funds, pairing sovereign capital with international matching grants.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "The National Climate Council has commenced disbursement of the initial tranche of sovereign ecological adaptation grants to sixteen vulnerable coastal and riverine states. Under the matching framework ratified during the inter-ministerial climate summit in Abuja, recipient states will direct liquidity toward concrete embankment construction, watershed stabilization, and automated early flood warning telemetry.",
				},
				{
					Type:    "quote",
					Quote:   "Resilience capital is no longer theoretical. Subnational governments now possess direct financing corridors to protect vulnerable agricultural and municipal belts before the peak seasonal rainfall.",
					Author:  "Dr. Salisu Dahiru",
					Context: "Director-General, National Council on Climate Change",
				},
				{
					Type: "paragraph",
					Text: "State governments in the Niger Delta and along the Benue River basin have established dedicated fiduciary trust accounts to ensure transparent procurement and real-time public tracking of flood barrier construction contracts.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "National Council on Climate Change",
					URL:          "https://natccc.gov.ng",
					PublishedAt:  "2026-10-14T10:00:00Z",
					ClaimSummary: "Subnational adaptation fund allocation guidelines.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "CLIMATE POLICY: 16 states receive first tranche of subnational climate adaptation capital for flood protection and watershed resilience 🌊🛡️",
				Facebook:   "The National Climate Council disburses matching grants for municipal flood defenses and coastal protection across 16 states.",
				LinkedIn:   "Ecological Governance: Subnational climate financing corridors operationalized across Nigeria.",
				Newsletter: "Policy Dispatch: First tranche of subnational adaptation funds released.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 17. Federal Road Safety Corps Mandates Digital Passenger Manifests for Commercial Corridors
		{
			ID:           deterministicUUID("interstate-transit-digital-manifest-enforcement"),
			Slug:         "interstate-transit-digital-manifest-enforcement",
			Title:        "Federal Road Safety Corps Mandates Digital Passenger Manifests for Commercial Corridors",
			Kicker:       "TRANSIT REGULATIONS · PUBLIC SAFETY",
			Deck:         "Interstate transport operators must migrate paper logbooks to real-time QR passenger verification systems by mid-quarter to ensure rapid passenger welfare coverage.",
			Category:     models.PillarCategory("practical-nigerian-updates"),
			CategoryName: "Practical Updates",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/2c9f5a31-9222-5444-8333-888899990000.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/2c9f5a31-9222-5444-8333-888899990000.webp",
				Caption:    "Digital ticketing scanners and manifest verification stations at intercity transit terminals.",
				Credit:     "FRSC Transport Safety Division",
			},
			ReadTime: "3 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Transport", "FRSC", "Public Safety", "Transit", "Digital Service"},
			Excerpt:  "Commercial motor park operators across all 36 states now face nationwide compliance checks on digital passenger manifests under newly gazetted safety regulations.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "The Federal Road Safety Corps (FRSC) has announced full enforcement timelines for the National Digital Passenger Manifest System (NDPMS), mandating commercial fleet operators to discard manual paper logs in favor of encrypted cloud manifests.",
				},
				{
					Type:    "quote",
					Quote:   "In the event of an emergency on federal highways, immediate digital identification and verified next-of-kin notification are essential legal rights for every commuter.",
					Author:  "Shehu Mohammed",
					Context: "Corps Marshal, FRSC",
				},
				{
					Type: "paragraph",
					Text: "Passengers booking intercity buses via registered terminals will now receive an automated SMS verification ticket containing route insurance data and direct emergency helpline links before vehicle departure.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Federal Road Safety Corps Gazette",
					URL:          "https://frsc.gov.ng",
					PublishedAt:  "2026-10-13T12:00:00Z",
					ClaimSummary: "Mandatory transit passenger manifest guidelines.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "TRANSIT UPDATE: FRSC mandates cloud passenger manifests for all commercial interstate buses. Paper registers phased out 🚌📋",
				Facebook:   "Commercial interstate commuters will now verify manifests digitally via SMS and QR codes under new FRSC safety rules.",
				LinkedIn:   "Transportation Policy: Digital manifest integration across commercial transit corridors.",
				Newsletter: "Practical Update: Cloud-verified passenger manifests mandated nationwide.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 18. Standards Organisation of Nigeria Issues Mandatory Safety Labels for Residential Solar Kits
		{
			ID:           deterministicUUID("son-mandates-component-certification-for-rooftop-solar"),
			Slug:         "son-mandates-component-certification-for-rooftop-solar",
			Title:        "Standards Organisation of Nigeria Issues Mandatory Safety Labels for Residential Solar Kits",
			Kicker:       "CONSUMER ADVISORY · ENERGY",
			Deck:         "Inverters, lithium iron phosphate battery packs, and photovoltaic panels must carry tamper-proof holographic verification seals starting next month.",
			Category:     models.PillarCategory("practical-nigerian-updates"),
			CategoryName: "Practical Updates",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/3d0a6b42-9333-5555-8444-999900001111.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/3d0a6b42-9333-5555-8444-999900001111.webp",
				Caption:    "Certified rooftop photovoltaic and battery inverter installation on a residential estate in Abuja.",
				Credit:     "SON Energy Standards Bureau",
			},
			ReadTime: "4 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Solar", "SON", "Energy", "Consumer Protection", "Utilities"},
			Excerpt:  "SON has unveiled tightened anti-counterfeiting benchmarks for renewable energy hardware to curb substandard battery imports and safeguard residential consumers.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "The Standards Organisation of Nigeria (SON) has directed all solar component distributors and importers to apply mandatory QR verification seals on inverters, battery management systems, and monocrystalline solar modules.",
				},
				{
					Type:    "quote",
					Quote:   "Nigerian households investing their hard-earned money in clean off-grid energy deserve genuine, tested components with verifiable warranty backing.",
					Author:  "Dr. Ifeanyi Okeke",
					Context: "Director-General, Standards Organisation of Nigeria",
				},
				{
					Type: "paragraph",
					Text: "Consumers can verify equipment authenticity directly at point of sale by scanning the tamper-evident holographic seal using the SON mobile app or SMS shortcode.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Standards Organisation of Nigeria Press Briefing",
					URL:          "https://son.gov.ng",
					PublishedAt:  "2026-10-12T14:30:00Z",
					ClaimSummary: "Rooftop solar and storage product certification standards.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "SOLAR SAFETY: SON mandates holographic QR verification seals on all inverters and lithium batteries to protect consumers from fake hardware ☀️🔋",
				Facebook:   "Buying a solar inverter or lithium battery? Check for the new SON holographic verification seal before paying.",
				LinkedIn:   "Clean Tech Standards: Regulatory compliance requirements for solar equipment importers in Nigeria.",
				Newsletter: "Consumer Advisory: How to verify your rooftop solar hardware under new SON rules.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 19. Institutional Investors Drive $450M Inflows into High-Yield Real Estate Investment Trusts
		{
			ID:           deterministicUUID("dual-listed-reits-unlock-commercial-property-liquidity"),
			Slug:         "dual-listed-reits-unlock-commercial-property-liquidity",
			Title:        "Institutional Investors Drive $450M Inflows into High-Yield Real Estate Investment Trusts",
			Kicker:       "CAPITAL MARKETS · REAL ESTATE",
			Deck:         "Grade-A commercial office towers and logistics warehouses across Victoria Island and Ikeja join dual-currency REIT platforms yielding inflation-hedged dividends.",
			Category:     models.PillarCategory("business-institutions"),
			CategoryName: "Business & Institutions",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/4e1b7c53-9444-5666-8555-000011112222.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/4e1b7c53-9444-5666-8555-000011112222.webp",
				Caption:    "Prime commercial property skyline along Victoria Island's financial boulevard.",
				Credit:     "Nigerian Exchange Group Media",
			},
			ReadTime: "5 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"REITs", "Real Estate", "NGX", "Investments", "Capital Markets"},
			Excerpt:  "Commercial real estate developers in Lagos and Abuja are securitizing warehouse and logistics footprints into high-performing investment trusts trading on the Nigerian Exchange.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "Institutional asset managers and pension fund administrators have committed over $450 million in newly structured Real Estate Investment Trusts (REITs) listed on the Nigerian Exchange, signaling renewed appetite for tangible commercial yield.",
				},
				{
					Type:    "quote",
					Quote:   "REIT securitization unlocks locked-up corporate balance sheet value while providing institutional funds with robust rental yields indexed to economic productivity.",
					Author:  "Temi Popoola",
					Context: "Group CEO, Nigerian Exchange Group",
				},
				{
					Type: "paragraph",
					Text: "The portfolio listings encompass cold-chain distribution centers in Sagamu, Grade-A financial towers in Lagos, and specialized fulfillment logistics centers catering to West Africa's e-commerce volume.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Nigerian Exchange Group Markets Bulletin",
					URL:          "https://ngxgroup.com",
					PublishedAt:  "2026-10-14T11:00:00Z",
					ClaimSummary: "Commercial REIT listings and capital flows report.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "REAL ESTATE FINANCE: $450M committed to commercial REITs on the NGX. Prime logistics and prime office portfolios lead the yield surge 🏢📈",
				Facebook:   "Commercial property securitization hits record milestone as pension funds channel capital into listed REITs.",
				LinkedIn:   "Capital Markets: Institutional liquidity flows into West African commercial property trusts.",
				Newsletter: "Markets Brief: How dual-listed REITs are unlocking real estate liquidity.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 20. PenCom Approves 25% RSA Balance Drawdown for First-Time Residential Mortgages
		{
			ID:           deterministicUUID("pencom-guidelines-allow-pension-equity-for-primary-mortgages"),
			Slug:         "pencom-guidelines-allow-pension-equity-for-primary-mortgages",
			Title:        "PenCom Approves 25% RSA Balance Drawdown for First-Time Residential Mortgages",
			Kicker:       "RETIREMENT REFORM · HOUSING",
			Deck:         "Thousands of salaried young professionals tap retirement savings accounts as verified primary equity contributions for long-term residential housing loans.",
			Category:     models.PillarCategory("business-institutions"),
			CategoryName: "Business & Institutions",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/5f2c8d64-9555-5777-8666-111122223333.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/5f2c8d64-9555-5777-8666-111122223333.webp",
				Caption:    "Newly completed master-planned residential housing development financed via pension-backed mortgages.",
				Credit:     "Federal Mortgage Bank Archive",
			},
			ReadTime: "4 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"PenCom", "Mortgage", "Pensions", "Housing", "Personal Finance"},
			Excerpt:  "PenCom's updated operational framework simplifies the process for retirement savings account holders to access up to a quarter of their balance for residential down payments.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "The National Pension Commission (PenCom) has streamlined regulatory prerequisites for RSA contributors seeking to utilize up to 25 percent of their accumulated pension balance as equity contribution for residential mortgage loans.",
				},
				{
					Type:    "quote",
					Quote:   "Housing security is a foundational pillar of retirement peace of mind. Allowing early equity access creates tangible lifetime wealth for contributors during their peak earning years.",
					Author:  "Omolola Oloworaran",
					Context: "Director-General, PenCom",
				},
				{
					Type: "paragraph",
					Text: "Over 34 primary mortgage institutions and commercial banks have integrated digital application verification portals directly with PenCom's contributor registry, reducing clearance turnaround from 90 days to under three weeks.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "National Pension Commission Circular",
					URL:          "https://pencom.gov.ng",
					PublishedAt:  "2026-10-11T09:00:00Z",
					ClaimSummary: "Operational guidelines on RSA balance drawdown for mortgage equity.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "HOUSING FINANCE: PenCom streamlines 25% pension equity access for first-time home buyers. Approval turnaround cut to 3 weeks 🏡🔑",
				Facebook:   "Did you know you can use up to 25% of your pension savings as down payment for your first home? Here is the new streamlined process.",
				LinkedIn:   "Retirement & Wealth: PenCom mortgage equity accessibility creates homeownership pathways for salaried professionals.",
				Newsletter: "Personal Finance: Guide to accessing your RSA balance for a primary mortgage.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 21. Midnight Calls Across Six Time Zones: The Intimate Geometry of Diaspora Romance
		{
			ID:           deterministicUUID("midnight-calls-across-six-time-zones-diaspora-love"),
			Slug:         "midnight-calls-across-six-time-zones-diaspora-love",
			Title:        "Midnight Calls Across Six Time Zones: The Intimate Geometry of Diaspora Romance",
			Kicker:       "PERSONAL ESSAY · MODERN LOVE",
			Deck:         "Between erratic transatlantic flight schedules, spotty WiFi, and divergent daylight hours, contemporary Nigerian couples redefine emotional closeness from London to Lagos.",
			Category:     models.PillarCategory("stories-romance"),
			CategoryName: "Stories & Romance",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/6a3d9e75-9666-5888-8777-222233334444.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/6a3d9e75-9666-5888-8777-222233334444.webp",
				Caption:    "Reflective evening shadows and shared phone conversations across transatlantic horizons.",
				Credit:     "Cherryl Media Essay Collection",
			},
			ReadTime: "5 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Modern Love", "Diaspora", "Relationships", "Memoir", "Essays"},
			Excerpt:  "A meditative exploration of long-distance courtship, FaceTime rituals, and the peculiar tenderness of building intimacy when oceans and time zones intervene.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "There is an unspoken geometry to loving someone who wakes up when you are preparing for bed. It begins with the instinctive calculation of hours—subtract one in winter, add five when summer daylight savings shift—and settles into a quiet cadence of whispered voice notes sent between tube stations and generator hums.",
				},
				{
					Type: "paragraph",
					Text: "For tens of thousands of young Nigerians who relocated in recent years, intimacy has transformed from shared evening meals in Ikeja into shared screen time, virtual cooking sessions, and the fragile vulnerability of transatlantic phone calls.",
				},
				{
					Type:    "quote",
					Quote:   "Distance doesn't diminish affection; it magnifies intention. Every minute of attention across oceans must be deliberately chosen.",
					Author:  "Bisi Akintola",
					Context: "Writer & Cultural Anthropologist",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Cherryl Media Cultural Archive",
					URL:          "https://cherrylmedia.com",
					PublishedAt:  "2026-10-14T18:00:00Z",
					ClaimSummary: "Modern diaspora interpersonal relationship essays.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "MODERN LOVE: Voice notes between tube stations and generator hums: the tender art of transatlantic Nigerian love across time zones ✈️❤️",
				Facebook:   "How modern Nigerian couples navigate long-distance relationships between Lagos, London, and Toronto without losing the spark.",
				LinkedIn:   "Culture & Society: The psychological and interpersonal dynamics of modern diaspora migration.",
				Newsletter: "Weekend Essay: Midnight calls across six time zones.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},

		// 22. Love on the Lekki Expressway: An Unlikely Serendipity in Gridlock
		{
			ID:           deterministicUUID("love-on-the-lekki-expressway-rush-hour-serendipity"),
			Slug:         "love-on-the-lekki-expressway-rush-hour-serendipity",
			Title:        "Love on the Lekki Expressway: An Unlikely Serendipity in Gridlock",
			Kicker:       "URBAN DISPATCH · ROMANCE",
			Deck:         "When an overheating radiator stalled two cars side by side at Chevron tollgate in peak evening traffic, an unexpected shared laugh turned into four years of marriage.",
			Category:     models.PillarCategory("stories-romance"),
			CategoryName: "Stories & Romance",
			Author: models.Author{
				Name: "",
				Role: "",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/7b4e0f86-9777-5999-8888-333344445555.webp",
				URL:        "https://pub-ceebfc5378214f48bb3d7ec7b50b0901.r2.dev/media/7b4e0f86-9777-5999-8888-333344445555.webp",
				Caption:    "Tail lights and evening glow reflecting off the wet asphalt along the Lekki-Epe expressway corridor.",
				Credit:     "Photo: Cherryl Media Field Archive",
			},
			ReadTime: "4 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Lagos", "Romance", "Urban Life", "City Stories", "Memoir"},
			Excerpt:  "A warm, witty chronicle of how Lagos traffic jams—ordinarily a source of collective exasperation—became the stage for one couple's serendipitous beginnings.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "In Lagos, the highway is rarely romantic. At 6:45 PM on a rainy Thursday near Chevron tollgate, it is typically a symphony of horns, brake lights, and roasted plantain vendors dodging side mirrors. Yet that evening, when two Japanese sedans suffered identical overheated radiators within ten feet of each other on the shoulder, circumstance engineered a meeting that neither dating apps nor dinner parties could have arranged.",
				},
				{
					Type:    "quote",
					Quote:   "If you can laugh together while pouring bottled water into a boiling car engine in the middle of Lagos traffic, you can survive almost anything together.",
					Author:  "Timi Balogun",
					Context: "Lekki Resident & Architect",
				},
				{
					Type: "paragraph",
					Text: "Four years later, the radiator water bottle still sits on their bookshelf in Victoria Island—a lighthearted testament to the spontaneous magic that thrives even in the city's most stubborn gridlock.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Cherryl Media Urban Memoirs",
					URL:          "https://cherrylmedia.com",
					PublishedAt:  "2026-10-13T19:00:00Z",
					ClaimSummary: "Lagos urban chronicles and personal essays.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:          "LAGOS STORIES: Overheating radiators at Chevron tollgate, shared bottled water, and 4 years of marriage. Proof that Lagos traffic has a sense of humor 💕🚗",
				Facebook:   "How a stalled car on the Lekki-Epe Expressway led to one couple's unexpected love story.",
				LinkedIn:   "Human Interest: The human stories and resilient humor of urban life in Lagos.",
				Newsletter: "Lagos Chronicles: Love on the Lekki Expressway.",
			},
			PublishedAt: now,
			CreatedAt:   now,
		},
	}
}
