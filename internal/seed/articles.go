package seed

import (
	"time"

	"github.com/google/uuid"
	"github.com/okoye-dev/cherryl-media-engine/internal/models"
)

// deterministicUUID generates a stable, reproducible UUID for an article slug.
func deterministicUUID(slug string) string {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("cherryl-article:"+slug)).String()
}

// GetAllWebsiteArticles compiles all canonical copy from the frontend (website/lib/articles.ts,
// homepage dispatches, narrative columns, and editorial dossiers) formatted as complete backend Article models.
func GetAllWebsiteArticles() []models.Article {
	now := time.Now().UTC()

	return []models.Article{
		// 1. Lead Story: Clean Energy Transition Accord
		{
			ID:           deterministicUUID("federal-executive-council-approves-unified-energy-framework"),
			Slug:         "federal-executive-council-approves-unified-energy-framework",
			Title:        "Federal Executive Council Confirms Unified National Clean Energy Transition Accord",
			Kicker:       "POLICY DEVELOPMENT · GOVERNANCE",
			Deck:         "Confirmed policy shifts introduce immediate tariff restructuring, regional mini-grid guarantees, and multi-state transmission oversight protocols.",
			Category:     models.CategoryCurrentAffairs,
			CategoryName: "Nigerian Current Affairs",
			Author: models.Author{
				Name: "Tunde Oladipo",
				Role: "Senior Policy Correspondent",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/articles/clean-energy-summit.webp",
				URL:        "/assets/images/clean-energy-summit.webp",
				Caption:    "Ministerial council convenes during the extraordinary session in Abuja.",
				Credit:     "Federal Information Bureau",
			},
			ReadTime: "4 min read",
			Status:   models.StatusPublished,
			IsLead:   true,
			Urgent:   true,
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
					Src:        "/assets/images/solar-substation.webp",
					StorageKey: "media/articles/solar-substation.webp",
					Caption:    "Regional mini-grid substation installation in Ogun State industrial corridor.",
					Credit:     "Photo: Cherryl Media Field Archive",
					Alt:        "Mini-grid solar distribution substation",
				},
				{
					Type:  "heading",
					Level: 2,
					Text:  "Fiscal Provisions and State Feeder Guarantees",
				},
				{
					Type: "paragraph",
					Text: "Crucially, the new framework mandates that distribution companies (DisCos) ring-fence industrial consumers on Band A feeders, ensuring that tariff pricing directly correlates with measured, metered uptime. Violations will trigger automated billing credits back to end consumers through unified metering software.",
				},
				{
					Type:     "internal_link",
					Title:    "NERC Releases Reclassified State Feeder Bands and Consumer Metering Schedule",
					Slug:     "nigeria-electricity-regulatory-commission-updates-feeder-bands",
					Category: "Practical Nigerian Updates",
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
			PublishedAt: now.Add(-3 * time.Hour),
			CreatedAt:   now.Add(-3 * time.Hour),
		},

		// 2. NERC Feeder Bands Consumer Advisory
		{
			ID:           deterministicUUID("nigeria-electricity-regulatory-commission-updates-feeder-bands"),
			Slug:         "nigeria-electricity-regulatory-commission-updates-feeder-bands",
			Title:        "NERC Releases Reclassified State Feeder Bands and Consumer Metering Schedule",
			Kicker:       "CONSUMER ADVISORY · PUBLIC SERVICE",
			Deck:         "Official public-service advisory detailing guaranteed supply hours, revised band thresholds, and direct customer dispute mechanisms.",
			Category:     models.CategoryPractical,
			CategoryName: "Practical Nigerian Updates",
			Author: models.Author{
				Name: "Bisi Alabi",
				Role: "Public Service & Utility Reporter",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/articles/metering-rollout.webp",
				URL:        "/assets/images/metering-rollout.webp",
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
					Level: 2,
					Text:  "What Changed for Residential and Commercial Consumers",
				},
				{
					Type: "paragraph",
					Text: "Under the revised metrics, Band A customers must receive a verified minimum of 20 hours daily. Any feeder that fails to maintain this threshold over three consecutive billing cycles must automatically be downgraded to Band B, with consumer billing adjusted downward immediately.",
				},
				{
					Type:    "quote",
					Quote:   "Consumers must not be billed for service tiers that their feeder lines do not physically deliver.",
					Author:  "Engr. Sanusi Garba",
					Context: "NERC Chairman",
				},
				{
					Type:      "external_link",
					Title:     "National Consumer Complaint Portal & Feeder Checker",
					URL:       "https://nerc.gov.ng/complaints",
					Publisher: "NERC Public Service Portal",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "NERC Commission Directives",
					URL:          "https://nerc.gov.ng",
					PublishedAt:  "2026-10-14T09:00:00Z",
					ClaimSummary: "Mandatory automated bill credits for feeder lines dropping below daily uptime thresholds.",
				},
			},
			// Notice: Optional social variants demo — only X & Facebook populated, LinkedIn/Newsletter omitted
			SocialVariants: models.SocialVariants{
				X:        "PRACTICAL UPDATE: NERC releases revised feeder band schedule. If your Band A feeder drops below 20 hrs/day for 3 days, DisCos must downgrade your tariff. Check your feeder status here ⚡",
				Facebook: "Important electricity update for Nigerian households: NERC has clarified the rules for Band A, B, and C feeders. DisCos are now required to issue automatic billing credits if daily hours fall below standard.",
			},
			PublishedAt: now.Add(-5 * time.Hour),
			CreatedAt:   now.Add(-5 * time.Hour),
		},

		// 3. Culture & Beatmakers Revolution
		{
			ID:           deterministicUUID("lagos-soundscapes-and-creators-revolution"),
			Slug:         "lagos-soundscapes-and-creators-revolution",
			Title:        "Inside the TikTok to Global Stage Pipeline: How Lagos Creators Reshape Pop Culture",
			Kicker:       "VIRAL CULTURE & MUSIC",
			Deck:         "From viral sound challenges to sold-out arenas: tracing how independent West African beatmakers turned internet culture into commercial leverage.",
			Category:     models.CategoryEntertainment,
			CategoryName: "Entertainment & Internet Culture",
			Author: models.Author{
				Name: "Tomiwa Adeleke",
				Role: "Culture & Trends Editor",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/articles/lagos-studio.webp",
				URL:        "/assets/images/lagos-studio.webp",
				Caption:    "Bedroom producer sessions transforming into global streaming hits.",
				Credit:     "Cherryl Media Sound Archive",
			},
			ReadTime: "5 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Afrobeats", "TikTok", "Music Business", "Lagos", "Creator Economy", "Pop Culture"},
			Excerpt:  "In small home studios across Yaba, Surulere, and Festac, twenty-year-old beatmakers are executing marketing campaigns that legacy record labels spend millions attempting to replicate.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "In small home studios across Yaba, Surulere, and Festac, twenty-year-old beatmakers are executing marketing campaigns that legacy record labels spend millions attempting to replicate. By designing 15-second loopable polyrhythms built specifically for TikTok dance challenges, these creators have reversed the traditional music distribution funnel.",
				},
				{
					Type:    "quote",
					Quote:   "We don't pitch radio stations anymore. The audience votes with 100,000 video creations before the full track is even uploaded to DSPs.",
					Author:  "P.Priime",
					Context: "Record Producer",
				},
				{
					Type:       "image",
					Src:        "/assets/images/afrobeats-crowd.webp",
					StorageKey: "media/articles/afrobeats-crowd.webp",
					Caption:    "Audience at a live showcase in Victoria Island singing along to unreleased viral snippets.",
					Credit:     "Photo: Cherryl Media",
					Alt:        "Concert crowd enjoying Afrobeats performance",
				},
				{
					Type: "paragraph",
					Text: "The financial architecture behind this shift is equally remarkable. Direct micropayments, digital brand partnerships, and independent distribution aggregator deals allow West African artists to retain up to 85% of their master rights.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Chart Data West Africa",
					URL:          "https://chartdata.org",
					PublishedAt:  "2026-10-13T15:30:00Z",
					ClaimSummary: "Streaming analytics tracking virality conversion rates from short-form audio clips.",
				},
			},
			SocialVariants: models.SocialVariants{
				X:        "How Lagos bedroom producers engineered a direct TikTok-to-O2 Arena pipeline without signing 360 label contracts. A deep dive into the business of modern African soundscapes 🎵🔥",
				Facebook: "Nollywood and Nigerian music are experiencing a distribution revolution. Here is how young producers in Lagos are bypassing traditional labels to create global hits directly on social platforms.",
			},
			PublishedAt: now.Add(-18 * time.Hour),
			CreatedAt:   now.Add(-18 * time.Hour),
		},

		// 4. Central Bank & AFC Local Currency Rails
		{
			ID:           deterministicUUID("cbn-leadership-announces-cross-border-fintech-settlement-rails"),
			Slug:         "cbn-leadership-announces-cross-border-fintech-settlement-rails",
			Title:        "Central Bank and AFC Formalize Unified Cross-Border Settlement Infrastructure",
			Kicker:       "FINTECH & INSTITUTIONS",
			Deck:         "Institutional announcement introduces local-currency trade settlement for commercial banks and registered payment switches.",
			Category:     models.CategoryBusiness,
			CategoryName: "Business & Institutions",
			Author: models.Author{
				Name: "Chukwudi Nwosu",
				Role: "Capital Markets Analyst",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/articles/afc-signing.webp",
				URL:        "/assets/images/afc-signing.webp",
				Caption:    "Central Bank leadership and AFC executives exchange bilateral agreements.",
				Credit:     "AFC Media Desk",
			},
			ReadTime: "4 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"CBN", "Fintech", "Banking", "Trade", "AFC", "CrossBorder", "Payments"},
			Excerpt:  "The Central Bank of Nigeria and the Africa Finance Corporation have jointly unveiled the Pan-African Local Currency Settlement Protocol (PALC-SP) to settle regional trade without third-party currency hops.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "The Central Bank of Nigeria and the Africa Finance Corporation have jointly unveiled the Pan-African Local Currency Settlement Protocol (PALC-SP). The system enables authorized Nigerian fintechs and commercial banks to clear regional cross-border transactions in domestic currencies, cutting trade reliance on foreign exchange correspondent accounts.",
				},
				{
					Type:    "quote",
					Quote:   "Intra-African commerce cannot expand if a trader in Lagos paying a supplier in Accra must route liquidity through New York or London.",
					Author:  "Olayemi Cardoso",
					Context: "CBN Governor",
				},
				{
					Type: "paragraph",
					Text: "The initial rollout integrates 14 commercial lenders and 6 licensed mobile money operators, with transaction settlement guaranteed in T+0 real time.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Central Bank Communications",
					URL:          "https://cbn.gov.ng",
					PublishedAt:  "2026-10-12T10:45:00Z",
					ClaimSummary: "Official communique on regional bilateral trade settlement integration.",
				},
			},
			// Social variant demo: only LinkedIn and X
			SocialVariants: models.SocialVariants{
				LinkedIn: "Strategic Development: The Central Bank of Nigeria and AFC have launched a local-currency trade settlement protocol. This significantly lowers transaction frictions for intra-African commerce and corporate treasuries.",
				X:        "INSTITUTIONAL WIRE: CBN & AFC formalize local-currency cross-border trade rails. Nigerian banks can now settle intra-African payments without dollar intermediary hops.",
			},
			PublishedAt: now.Add(-28 * time.Hour),
			CreatedAt:   now.Add(-28 * time.Hour),
		},

		// 5. Romance & Transit Essay: Bridge at Dusk
		{
			ID:           deterministicUUID("conversations-on-third-mainland-bridge-at-dusk"),
			Slug:         "conversations-on-third-mainland-bridge-at-dusk",
			Title:        "Conversations on the Bridge at Dusk: Modern Love, Hustle, and Lagos Distance",
			Kicker:       "ESSAYS & MODERN LOVE",
			Deck:         "A personal narrative on romance across the mainland-island divide, unspoken compromises, and the quiet beauty of evening transit.",
			Category:     models.CategoryRomance,
			CategoryName: "Stories & Romance",
			Author: models.Author{
				Name: "Amara Okonkwo",
				Role: "Special Features Columnist",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/articles/lagos-dusk.webp",
				URL:        "/assets/images/lagos-dusk.webp",
				Caption:    "Brake lights reflecting over the lagoon as twilight settles over Lagos.",
				Credit:     "Urban Lens Archive",
			},
			ReadTime: "6 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Lagos", "Romance", "Essay", "CityLife", "Transit", "Relationships"},
			Excerpt:  "In Lagos, geography is the unspoken third party in every relationship. To love someone who lives in Ikeja when your office and bedroom are in Lekki is to accept that romance is measured in kilometers of asphalt.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "In Lagos, geography is the unspoken third party in every relationship. To love someone who lives in Ikeja when your office and bedroom are in Lekki Phase 1 is to accept that romance will be measured not in flowers or dinners, but in kilometers of asphalt and the temperament of the Third Mainland Bridge.",
				},
				{
					Type: "paragraph",
					Text: "We learned to time our conversations by the movement of commuter yellow buses. 'I'm at Adekunle,' meant you had eight minutes before the signal dipped. 'I just hit Adeniji,' meant relief was imminent.",
				},
				{
					Type:   "quote",
					Quote:  "We built an entire vocabulary out of traffic reports and voice notes sent during standstill gridlock.",
					Author: "Amara Okonkwo",
				},
				{
					Type: "paragraph",
					Text: "Yet, when the sun sinks into the lagoon and the evening sky turns a bruised shade of violet and terracotta, there is a singular tenderness to this city. You realize that endurance is itself a language of care.",
				},
			},
			Sources: []models.SourceCitation{
				{
					SourceName:   "Cherryl Stories & Culture Bureau",
					URL:          "https://cherrylmedia.com/stories",
					PublishedAt:  "2026-10-11T17:00:00Z",
					ClaimSummary: "First-person narrative essay commissioned for Cherryl Weekend Dossier.",
				},
			},
			SocialVariants: models.SocialVariants{
				Newsletter: "Weekend Essay: 'Conversations on the Bridge at Dusk' — Amara Okonkwo on modern love, long commutes, and finding tenderness in the world's most relentless city.",
				X:          "'In Lagos, geography is the unspoken third party in every relationship.' A Saturday narrative on modern courtship, distance, and evening transit ❤️📖",
			},
			PublishedAt: now.Add(-40 * time.Hour),
			CreatedAt:   now.Add(-40 * time.Hour),
		},

		// 6. Homepage Featured Dispatch: Ministry of Transport Safety Standards
		{
			ID:           deterministicUUID("federal-ministry-of-transport-finalizes-safety-standard"),
			Slug:         "federal-ministry-of-transport-finalizes-safety-standard",
			Title:        "Federal Ministry of Transport Finalizes Comprehensive Inter-State Safety Standard",
			Kicker:       "TRANSPORT SAFETY & STANDARDS",
			Deck:         "Federal inter-state transit safety code mandates speed governors, digital passenger manifests, and roadworthiness certification across all interstate transport terminals.",
			Category:     models.CategoryCurrentAffairs,
			CategoryName: "Nigerian Current Affairs",
			Author: models.Author{
				Name: "Bisi Alabi",
				Role: "Public Service & Utility Reporter",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/articles/transport-safety.webp",
				URL:        "https://images.unsplash.com/photo-1544620347-c4fd4a3d5957?auto=format&fit=crop&q=80&w=1200",
				Caption:    "Interstate passenger coaches undergoing safety compliance checks.",
				Credit:     "Federal Road Safety Corps Archive",
			},
			ReadTime: "3 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Transport", "Safety", "FRSC", "Interstate", "Policy", "Mobility"},
			Excerpt:  "The Federal Ministry of Transportation has gazetted mandatory digital manifests and automated speed telemetry for all commercial inter-state fleet operators.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "The Federal Ministry of Transportation has gazetted a comprehensive road safety guideline requiring commercial transport operators to implement synchronized digital passenger manifests and onboard speed limit telemetry by the first quarter of next year.",
				},
				{
					Type: "paragraph",
					Text: "The policy, developed in conjunction with the Federal Road Safety Corps (FRSC), aims to eliminate unmanifested passenger boarding and curtail highway accidents on major transit corridors including the Lagos-Ibadan expressway and Abuja-Kaduna highway.",
				},
			},
			PublishedAt: now.Add(-8 * time.Hour),
			CreatedAt:   now.Add(-8 * time.Hour),
		},

		// 7. Homepage Practical Advisory: Phase 2 Direct Smart Meter Rollout
		{
			ID:           deterministicUUID("distribution-companies-commence-phase-2-direct-meter-rollout"),
			Slug:         "distribution-companies-commence-phase-2-direct-meter-rollout",
			Title:        "Practical Advisory: Distribution Companies Commence Phase 2 Direct Smart Meter Rollout",
			Kicker:       "PRACTICAL ADVISORY",
			Deck:         "Distribution companies deploy free prepaid smart meters across residential clusters to end estimated billing disputes and improve billing transparency.",
			Category:     models.CategoryPractical,
			CategoryName: "Practical Nigerian Updates",
			Author: models.Author{
				Name: "Bisi Alabi",
				Role: "Public Service & Utility Reporter",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/articles/smart-metering.webp",
				URL:        "https://images.unsplash.com/photo-1509390144018-eeaf65052242?auto=format&fit=crop&q=80&w=1200",
				Caption:    "Prepaid digital smart meter installation in residential estate.",
				Credit:     "Cherryl Media Staff File",
			},
			ReadTime: "4 min read",
			Status:   models.StatusPublished,
			Tags:     []string{"Metering", "Power", "Electricity", "Consumer", "DisCos", "Utilities"},
			Excerpt:  "Electricity distribution companies have commenced Phase 2 of the National Mass Metering Programme, prioritizing high-density unmetered residential corridors.",
			Body: []models.ContentBlock{
				{
					Type:      "paragraph",
					IsDropCap: true,
					Text:      "Electricity distribution companies nationwide have commenced Phase 2 of the National Mass Metering Programme, deploying over 450,000 smart prepayment meters across designated urban and suburban feeder clusters.",
				},
				{
					Type: "paragraph",
					Text: "Customers within active installation corridors will receive notifications via SMS detailing verification documents required, with DisCos reiterating that meter units are provided at zero upfront cost under the federal assistance facility.",
				},
			},
			PublishedAt: now.Add(-12 * time.Hour),
			CreatedAt:   now.Add(-12 * time.Hour),
		},

		// 8. Capital Markets: Sovereign Local-Currency Trade Facilities
		{
			ID:           deterministicUUID("commercial-lenders-expand-sovereign-local-currency-trade-facilities"),
			Slug:         "commercial-lenders-expand-sovereign-local-currency-trade-facilities",
			Title:        "Commercial Lenders Expand Sovereign Local-Currency Trade Facilities",
			Kicker:       "CAPITAL MARKETS",
			Deck:         "Commercial banking syndicate structures N500bn trade financing liquidity pool for manufacturing exporters and agro-allied processors.",
			Category:     models.CategoryBusiness,
			CategoryName: "Business & Institutions",
			Author: models.Author{
				Name: "Chukwudi Nwosu",
				Role: "Capital Markets Analyst",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/articles/trade-finance.webp",
				URL:        "https://images.unsplash.com/photo-1590283603385-17ffb3a7f29f?auto=format&fit=crop&q=80&w=1200",
				Caption:    "Commercial banking trading desk in Victoria Island, Lagos.",
				Credit:     "Reuters / Financial Markets Archive",
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
				{
					Type: "paragraph",
					Text: "The liquidity mechanism allows export-oriented small and medium enterprises to access single-digit interest financing, insulating their operations from external dollar volatility while expanding regional trade footprint.",
				},
			},
			PublishedAt: now.Add(-16 * time.Hour),
			CreatedAt:   now.Add(-16 * time.Hour),
		},

		// 9. Stories & Romance: Letters on Awolowo Road
		{
			ID:           deterministicUUID("letters-on-awolowo-road-and-the-intimacy-of-modern-distance"),
			Slug:         "letters-on-awolowo-road-and-the-intimacy-of-modern-distance",
			Title:        "Letters on Awolowo Road: The Intimacy of Modern Distance in Ikoyi",
			Kicker:       "STORIES & ROMANCE",
			Deck:         "An evocative personal reflection on courtship, late-night dinners, and navigating emotional distance in Lagos.",
			Category:     models.CategoryRomance,
			CategoryName: "Stories & Romance",
			Author: models.Author{
				Name: "Amara Okonkwo",
				Role: "Special Features Columnist",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/articles/awolowo-road.webp",
				URL:        "https://images.unsplash.com/photo-1518895949257-7621c3c786d7?auto=format&fit=crop&q=80&w=1200",
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
				{
					Type: "paragraph",
					Text: "In Lagos, finding someone who listens to you without checking their phone every ninety seconds feels less like courtship and more like a quiet miracle.",
				},
			},
			PublishedAt: now.Add(-22 * time.Hour),
			CreatedAt:   now.Add(-22 * time.Hour),
		},

		// 10. Dossier: Inside the Lagos Underground Afro-Fusion
		{
			ID:           deterministicUUID("inside-the-lagos-underground-polyphonic-afro-fusion"),
			Slug:         "inside-the-lagos-underground-polyphonic-afro-fusion",
			Title:        "Inside the Lagos Underground: The Sound Architects Engineering Polyphonic Afro-Fusion",
			Kicker:       "SOUNDSCAPES & ACOUSTICS",
			Deck:         "A technical exploration into the rhythmic syncopations and modular synthesizers defining new West African acoustics.",
			Category:     models.CategoryEntertainment,
			CategoryName: "Entertainment & Internet Culture",
			Author: models.Author{
				Name: "Funke Balogun",
				Role: "Culture & Entertainment Bureau Chief",
			},
			FeaturedImage: models.FeaturedImage{
				StorageKey: "media/articles/afro-fusion-studio.webp",
				URL:        "https://images.unsplash.com/photo-1598488035139-bdbb2231ce04?auto=format&fit=crop&q=80&w=1200",
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
				{
					Type: "paragraph",
					Text: "Rather than copying Western pop structures, these producers build textured, hypnotic soundscapes that respect the syncopated spaces between percussion strokes.",
				},
			},
			PublishedAt: now.Add(-30 * time.Hour),
			CreatedAt:   now.Add(-30 * time.Hour),
		},
	}
}
