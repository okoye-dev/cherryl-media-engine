package models

import "time"

type PillarCategory string

const (
	CategoryStateAffairs            PillarCategory = "state-affairs-environs"
	CategoryNewsbeat                PillarCategory = "newsbeat"
	CategoryBrandsSMEs              PillarCategory = "brands-smes"
	CategoryOpinions                PillarCategory = "opinions"
	CategorySports                  PillarCategory = "sports"
	CategoryEntertainmentLifestyle  PillarCategory = "entertainment-lifestyle"
	CategoryEducation               PillarCategory = "education"
	CategoryThrillsTricks           PillarCategory = "thrills-tricks"
	CategoryYouthCorner             PillarCategory = "youth-corner"
	CategoryTourHotels              PillarCategory = "tour-hotels"

	// Legacy aliases for backward compatibility
	CategoryCurrentAffairs PillarCategory = "nigerian-current-affairs"
	CategoryEntertainment  PillarCategory = "entertainment-internet-culture"
	CategoryPractical      PillarCategory = "practical-nigerian-updates"
	CategoryBusiness       PillarCategory = "business-institutions"
	CategoryRomance        PillarCategory = "stories-romance"
)

type PublishingStatus string

const (
	StatusDraft     PublishingStatus = "draft"
	StatusInReview  PublishingStatus = "in_review"
	StatusApproved  PublishingStatus = "approved"
	StatusPublished PublishingStatus = "published"
)

type ContentBlock struct {
	ID         string `json:"id,omitempty"`
	OrderID    int    `json:"order_id,omitempty"`
	Type       string `json:"type"`
	Text       string `json:"text,omitempty"`
	IsDropCap  bool   `json:"isDropCap,omitempty"`
	Level      int    `json:"level,omitempty"`
	Quote      string `json:"quote,omitempty"`
	Author     string `json:"author,omitempty"`
	Context    string `json:"context,omitempty"`
	StorageKey string `json:"storage_key,omitempty"`
	Src        string `json:"src,omitempty"`
	Alt        string `json:"alt,omitempty"`
	Caption    string `json:"caption,omitempty"`
	Credit     string `json:"credit,omitempty"`
	Title      string `json:"title,omitempty"`
	URL        string `json:"url,omitempty"`
	Slug       string `json:"slug,omitempty"`
	Category   string `json:"category,omitempty"`
	Deck       string `json:"deck,omitempty"`
	Publisher  string `json:"publisher,omitempty"`
	IsInternal bool   `json:"isInternal,omitempty"`
	VideoURL   string `json:"videoUrl,omitempty"`
	ReadTime   string `json:"readTime,omitempty"`
	Verified   bool   `json:"verified,omitempty"`
	ImageURL   string `json:"imageUrl,omitempty"`
}

type Author struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type FeaturedImage struct {
	StorageKey string `json:"storage_key,omitempty"`
	URL        string `json:"url"`
	Caption    string `json:"caption,omitempty"`
	Credit     string `json:"credit,omitempty"`
}

type SocialVariants struct {
	X          string `json:"x,omitempty"`
	Facebook   string `json:"facebook,omitempty"`
	LinkedIn   string `json:"linkedin,omitempty"`
	Newsletter string `json:"newsletter,omitempty"`
}

type SourceCitation struct {
	SourceName   string `json:"sourceName"`
	URL          string `json:"url"`
	PublishedAt  string `json:"publishedAt,omitempty"`
	ClaimSummary string `json:"claimSummary,omitempty"`
}

type Source struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Type             string    `json:"type"`
	FeedURL          string    `json:"feedUrl"`
	SiteURL          string    `json:"siteUrl,omitempty"`
	CategoryHint     string    `json:"categoryHint,omitempty"`
	ReliabilityScore float64   `json:"reliabilityScore"`
	Active           bool      `json:"active"`
	LastFetchedAt    time.Time `json:"lastFetchedAt,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
}

type SuggestedStory struct {
	ID             string           `json:"id"`
	Headline       string           `json:"headline"`
	Pitch          string           `json:"pitch"`
	Category       string           `json:"category"`
	RelevanceScore float64          `json:"relevanceScore"`
	Sources        []SourceCitation `json:"sources"`
	ExtractedFacts []string         `json:"extractedFacts"`
	SuggestedDraft []ContentBlock   `json:"suggestedDraft,omitempty"`
	Status         string           `json:"status"` // 'discovered', 'selected', 'drafted', 'in_review', 'rejected', 'published'
	ArticleID      string           `json:"articleId,omitempty"`
	CreatedAt      time.Time        `json:"createdAt"`
}

type Article struct {
	ID             string           `json:"id"`
	Slug           string           `json:"slug"`
	Title          string           `json:"title"`
	Subtitle       string           `json:"subtitle,omitempty"`
	Kicker         string           `json:"kicker,omitempty"`
	Deck           string           `json:"deck,omitempty"`
	Category       PillarCategory   `json:"category"`
	CategoryName   string           `json:"categoryName,omitempty"`
	Author         Author           `json:"author"`
	FeaturedImage  FeaturedImage    `json:"featuredImage"`
	ReadTime       string           `json:"readTime,omitempty"`
	Status         PublishingStatus `json:"status"`
	Tags           []string         `json:"tags,omitempty"`
	Excerpt        string           `json:"excerpt,omitempty"`
	OriginalURL    string           `json:"originalUrl,omitempty"`
	IsLead         bool             `json:"isLead,omitempty"`
	Urgent         bool             `json:"urgent,omitempty"`
	IsFeatured     bool             `json:"isFeatured,omitempty"`
	MediaConfirmed bool             `json:"mediaConfirmed,omitempty"`
	SendWebpushr   bool             `json:"send_webpushr,omitempty"`
	SEOTitle       string           `json:"seo_title,omitempty"`
	SEOSlug        string           `json:"seo_slug,omitempty"`
	SEODescription string           `json:"seo_description,omitempty"`
	Body           []ContentBlock   `json:"body"`
	RelatedStories []ContentBlock   `json:"related_stories,omitempty"`
	Sources        []SourceCitation `json:"sources,omitempty"`
	SocialVariants SocialVariants   `json:"socialVariants,omitempty"`
	CandidateID    string           `json:"candidateId,omitempty"`
	AuthorID       string           `json:"authorId,omitempty"`
	PublishedBy    string           `json:"publishedBy,omitempty"`
	AIReviewNotes  string           `json:"aiReviewNotes,omitempty"`
	PublishedAt    time.Time        `json:"publishedAt"`
	CreatedAt      time.Time        `json:"createdAt"`
}
