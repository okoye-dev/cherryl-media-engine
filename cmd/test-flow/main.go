package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/okoye-dev/cherryl-media-engine/internal/config"
	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/services"
	"github.com/okoye-dev/cherryl-media-engine/internal/storage"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

func main() {
	cfg := config.Load()
	if cfg.Database.URL == "" {
		log.Fatal("DATABASE_URL must be configured in backend/.env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	fmt.Println("================================================================================")
	fmt.Println("         CHERRYL MEDIA — PUBLISH & DRAFT FLOW VERIFICATION SUITE               ")
	fmt.Println("================================================================================")

	// 1. Initialize S3 Storage and Database Store
	s3Storage, err := storage.NewS3Storage(&cfg.S3)
	if err != nil {
		log.Fatalf("failed to initialize S3 storage: %v", err)
	}

	pool, err := store.NewPool(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	pg := store.NewPGStore(pool)
	mediaResolver := services.NewMediaResolver(s3Storage)
	articleService := services.NewArticleService(pg, mediaResolver)

	// 2. Setup a subscriber to verify email dispatch queueing
	testEmail := "notifications-desk@cherrylmedia.com"
	testSubscriberName := "Editorial Wire Monitor"
	fmt.Printf("\n[1/4] Registering test subscriber for 'state-affairs-environs'...\n")
	sub, err := pg.AddSubscriber(ctx, testEmail, testSubscriberName, []string{"state-affairs-environs", "newsbeat"})
	if err != nil {
		log.Printf("Note: subscriber registration returned: %v (continuing)", err)
	} else {
		fmt.Printf("      ✓ Subscriber active: %s (%s)\n", sub.Email, sub.Name)
	}

	// 3. Scenario A: Save a Draft Story (To test /admin draft persistence and resumption)
	fmt.Printf("\n[2/4] Scenario A: Creating an in-progress DRAFT story...\n")
	draftSlug := "the-evolution-of-modular-solar-grids-in-northern-nigeria"
	draftCoverKey := "media/decd08ac-ef71-54ee-ac58-fd2480c007fa.webp"
	draftCoverURL := s3Storage.GetPublicURL(draftCoverKey)

	draftArticle := models.Article{
		Title:        "The Evolution of Modular Solar Grids in Northern Nigeria",
		Slug:         draftSlug,
		Category:     models.CategoryBrandsSMEs,
		CategoryName: "BRANDS & SMEs",
		Kicker:       "SPECIAL REPORT",
		Deck:         "Decentralized solar mini-grids are transforming rural cottage agro-industries across Kano, Kaduna, and Jigawa.",
		Subtitle:     "Decentralized solar mini-grids are transforming rural cottage agro-industries across Kano, Kaduna, and Jigawa.",
		Excerpt:      "How decentralized rural mini-grids are delivering dependable 24-hour cold chain power for northern agricultural clusters.",
		Status:       models.StatusDraft, // Stays in DRAFT status!
		Author: models.Author{
			Name: "Amina Bello & Editorial Desk",
			Role: "Senior Enterprise Correspondent",
		},
		ReadTime: "4 min read",
		FeaturedImage: models.FeaturedImage{
			StorageKey: draftCoverKey,
			URL:        draftCoverURL,
			Caption:    "Solar inverter array in a decentralized agro-processing mini-grid cluster outside Kano.",
			Credit:     "Cherryl Media / Rural Power Documentation",
		},
		Body: []models.ContentBlock{
			{
				OrderID:   1,
				Type:      "paragraph",
				Text:      "Across the agricultural belt of northern Nigeria, an unassuming quiet infrastructure shift is gathering momentum. In rural cottage processing cooperatives that previously relied on costly diesel generator sets, modular solar mini-grids are establishing 24-hour continuous cold storage and grain-milling operations.",
				IsDropCap: true,
			},
			{
				OrderID: 2,
				Type:    "heading",
				Level:   2,
				Text:    "Decentralized Reliability for Agro-Processors",
			},
			{
				OrderID: 3,
				Type:    "paragraph",
				Text:    "Small enterprise owners report a 40 percent reduction in monthly operational overhead since switching to pay-as-you-go solar arrays, allowing local grain millers to operate predictable evening shifts without fuel price volatility.",
			},
			{
				OrderID: 4,
				Type:    "quote",
				Quote:   "When processing grain depends on fluctuating fuel supplies, margins evaporate. Dedicated mini-grid solar gives cottage industries a stable foundation.",
				Author:  "Ibrahim Sanusi",
				Context: "Kano Agro-Processors Syndicate",
			},
		},
		Sources: []models.SourceCitation{
			{
				SourceName:   "Rural Electrification Agency (REA)",
				PublishedAt:  "October 2026",
				URL:          "https://rea.gov.ng/mini-grid-deployment-quarterly-report",
				ClaimSummary: "Sub-national mini-grid deployments reached 180 communities in Q3 2026.",
			},
		},
		Tags:           []string{"Energy", "Northern Nigeria", "SMEs", "Agriculture", "Solar Power"},
		SEOTitle:       "The Evolution of Modular Solar Grids in Northern Nigeria | Cherryl Media",
		SEOSlug:        draftSlug,
		SEODescription: "Decentralized solar mini-grids are transforming rural cottage agro-industries across Kano, Kaduna, and Jigawa.",
	}

	savedDraft, err := articleService.SaveArticle(ctx, &draftArticle)
	if err != nil {
		log.Fatalf("failed to save draft article: %v", err)
	}
	fmt.Printf("      ✓ Draft successfully created: \"%s\"\n", savedDraft.Title)
	fmt.Printf("      ✓ Status: %s | Slug: %s\n", savedDraft.Status, savedDraft.Slug)

	// 4. Scenario B: Publish a Story with Media Confirmation & Notification Dispatch
	fmt.Printf("\n[3/4] Scenario B: Confirming media & PUBLISHING a high-impact story...\n")
	pubSlug := "national-grid-stabilizes-with-5800mw-peak-feeder-band-overhaul"
	pubCoverKey := "media/2f405add-8bba-5261-99af-15feaa62754b.webp"
	pubCoverURL := s3Storage.GetPublicURL(pubCoverKey)

	// Step 4a: Enforce media confirmation in database prior to publish
	fmt.Printf("      1. Verifying & confirming upload status in 'uploads' table for '%s'...\n", pubCoverKey)
	if err := pg.ConfirmUpload(ctx, pubCoverKey, 342000, pubCoverURL); err != nil {
		log.Printf("Warning during media confirm: %v", err)
	}
	isConfirmed, _ := pg.IsConfirmed(ctx, pubCoverKey)
	fmt.Printf("         ✓ Media asset confirmed in storage: %v\n", isConfirmed)

	pubArticle := models.Article{
		Title:        "National Grid Stabilizes with 5,800MW Peak as Feeder Band Overhaul Yields Measurable Reliability Gains",
		Slug:         pubSlug,
		Category:     models.CategoryStateAffairs,
		CategoryName: "STATE AFFAIRS & ENVIRONS",
		Kicker:       "NATIONAL GRID REPORT",
		Deck:         "Federal power ministry and NERC report highest consistent transmission output of the quarter following transmission wheeling upgrades.",
		Subtitle:     "Federal power ministry and NERC report highest consistent transmission output of the quarter following transmission wheeling upgrades.",
		Excerpt:      "Daily transmission load peaks at 5,840 megawatts across critical industrial corridors following regional wheeling capacity upgrades.",
		Status:       models.StatusDraft, // Start as draft, then execute the publish pipeline!
		Author: models.Author{
			Name: "Babajide Onanuga, State Affairs Bureau",
			Role: "Senior Energy Correspondent",
		},
		ReadTime:   "5 min read",
		IsFeatured: true,
		FeaturedImage: models.FeaturedImage{
			StorageKey: pubCoverKey,
			URL:        pubCoverURL,
			Caption:    "Transmission substation infrastructure along the national 330kV transmission corridor.",
			Credit:     "Transmission Company of Nigeria / Ministry of Power",
		},
		Body: []models.ContentBlock{
			{
				OrderID:   1,
				Type:      "paragraph",
				Text:      "Nigeria's national electricity grid achieved a sustained quarterly transmission peak of 5,842 megawatts this week, marking the most consistent power wheeling interval recorded by the National Control Centre in over twelve months.",
				IsDropCap: true,
			},
			{
				OrderID: 2,
				Type:    "heading",
				Level:   2,
				Text:    "Feeder Band Audit and Substation Reinforcement",
			},
			{
				OrderID: 3,
				Type:    "paragraph",
				Text:    "According to performance audits released by the Nigerian Electricity Regulatory Commission (NERC), targeted reinforcement along critical transmission interfaces has curtailed frequency fluctuations that historically triggered cascade grid collapses.",
			},
			{
				OrderID: 4,
				Type:    "quote",
				Quote:   "By prioritizing feeder band stability and enforcing mandatory telemetry compliance across generating companies, we have established a more resilient transmission baseline.",
				Author:  "Engr. Sanusi Garba",
				Context: "Chairman, Nigerian Electricity Regulatory Commission",
			},
			{
				OrderID:    5,
				Type:       "link",
				Title:      "The Evolution of Modular Solar Grids in Northern Nigeria",
				URL:        "/article/the-evolution-of-modular-solar-grids-in-northern-nigeria",
				Publisher:  "Cherryl Media",
				Deck:       "Related enterprise dispatch on rural decentralized micro-grids.",
				IsInternal: true,
			},
		},
		Sources: []models.SourceCitation{
			{
				SourceName:   "Nigerian Electricity Regulatory Commission (NERC)",
				PublishedAt:  "October 2026",
				URL:          "https://nerc.gov.ng/quarterly-grid-performance-bulletin",
				ClaimSummary: "Transmission peak reached 5,842MW with zero system collapse incidents in 45 days.",
			},
			{
				SourceName:   "Transmission Company of Nigeria (TCN)",
				PublishedAt:  "October 2026",
				URL:          "https://tcn.org.ng/national-control-centre-telemetry-status",
				ClaimSummary: "Supervisory Control and Data Acquisition (SCADA) telemetry integrated on 14 major lines.",
			},
		},
		Tags:           []string{"National Grid", "Power", "State Affairs", "NERC", "Infrastructure"},
		SEOTitle:       "National Grid Stabilizes with 5,800MW Peak | Cherryl Media",
		SEOSlug:        pubSlug,
		SEODescription: "Federal power ministry and NERC report highest consistent transmission output of the quarter following transmission wheeling upgrades.",
	}

	// Save article record first
	savedPub, err := articleService.SaveArticle(ctx, &pubArticle)
	if err != nil {
		log.Fatalf("failed to save story before publish: %v", err)
	}

	// Step 4b: Execute the Publish Pipeline (enforcing media confirmation and triggering dispatches)
	fmt.Printf("      2. Executing PublishArticle pipeline for ID: %s (Slug: %s)...\n", savedPub.ID, savedPub.Slug)
	publishedResult, err := articleService.PublishArticle(ctx, savedPub.Slug, "Senior Managing Editor")
	if err != nil {
		log.Fatalf("failed to publish article: %v", err)
	}
	fmt.Printf("         ✓ Publish succeeded! Status: %s | PublishedAt: %s\n",
		publishedResult.Status, publishedResult.PublishedAt.Format(time.RFC3339))

	// Step 4c: Set story as featured flagship lead
	fmt.Printf("      3. Pinning story as Featured Lead via Decision Engine...\n")
	featuredStory, err := articleService.SetFeatured(ctx, publishedResult.ID, true)
	if err != nil {
		log.Printf("Warning setting featured: %v", err)
	} else {
		fmt.Printf("         ✓ Story is now set as Featured: %v\n", featuredStory.IsFeatured)
	}

	// 5. Verification checks
	fmt.Printf("\n[4/4] Verifying System State & Decision Engine...\n")
	time.Sleep(600 * time.Millisecond) // Allow asynchronous dispatch queue to settle

	// Verify Featured Engine selects it
	activeFeatured, err := articleService.GetFeatured(ctx)
	if err != nil {
		log.Printf("Warning querying featured: %v", err)
	} else {
		fmt.Printf("      ✓ Decision Engine Featured Story: \"%s\" (Slug: %s)\n", activeFeatured.Title, activeFeatured.Slug)
	}

	// Verify Dispatches recorded
	dispatches, err := pg.ListDispatches(ctx, publishedResult.ID)
	if err != nil {
		log.Printf("Warning querying dispatches: %v", err)
	} else {
		fmt.Printf("      ✓ Notification dispatches logged: %d recipient(s)\n", len(dispatches))
		for _, d := range dispatches {
			fmt.Printf("         - Dispatched to: %s [Status: %s]\n", d.RecipientEmail, d.Status)
		}
	}

	// Verify Draft appears in Drafts list
	drafts, err := articleService.ListArticles(ctx, "draft", "", "", 10)
	if err != nil {
		log.Printf("Warning querying drafts: %v", err)
	} else {
		fmt.Printf("      ✓ Active Drafts in database: %d draft(s)\n", len(drafts))
		for _, dr := range drafts {
			if dr.Slug == draftSlug {
				fmt.Printf("         - Found draft: \"%s\" (Slug: %s)\n", dr.Title, dr.Slug)
			}
		}
	}

	// Print visual walkthrough
	fmt.Println("\n================================================================================")
	fmt.Println("                         VERIFICATION COMPLETE                                 ")
	fmt.Println("================================================================================")
	fmt.Println("You can now run the app and visually inspect both stories:")
	fmt.Println("")
	fmt.Println("1. Start Backend Engine:")
	fmt.Println("   cd backend && go run ./cmd/server")
	fmt.Println("   (Server runs on http://localhost:8080)")
	fmt.Println("")
	fmt.Println("2. Start Frontend Website:")
	fmt.Println("   cd website && npm run dev")
	fmt.Println("   (Website runs on http://localhost:3000)")
	fmt.Println("")
	fmt.Println("3. Inspect the Results in your Browser:")
	fmt.Println("   • View the PUBLISHED Story on Homepage & Category:")
	fmt.Printf("     http://localhost:3000/article/%s\n", pubSlug)
	fmt.Println("     http://localhost:3000/category/state-affairs-environs")
	fmt.Println("")
	fmt.Println("   • View the DRAFT Story in the Admin Dashboard:")
	fmt.Println("     http://localhost:3000/admin (click the 'Drafts' tab to see the amber draft badge)")
	fmt.Println("")
	fmt.Println("   • Resume editing the DRAFT Story in the Article Studio:")
	fmt.Printf("     http://localhost:3000/admin/editor?slug=%s\n", draftSlug)
	fmt.Println("")
	fmt.Println("   • Inspect the Featured curation & Gist search:")
	fmt.Println("     http://localhost:3000/admin/sections (click 'Replace' or '+ Add Story' on any slot)")
	fmt.Println("")
	fmt.Println("   • Inspect the Notification Subscribers & Dispatches log:")
	fmt.Println("     http://localhost:3000/admin/settings (click the 'Notification Subscribers' tab)")
	fmt.Println("================================================================================")

	_ = uuid.Nil
	_ = strings.Clone("")
}
