package models

import "time"

type PillarCategory string

const (
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
	Type       string `json:"type"`
	Text       string `json:"text,omitempty"`
	IsDropCap  bool   `json:"isDropCap,omitempty"`
	Level      int    `json:"level,omitempty"`
	Quote      string `json:"quote,omitempty"`
	Author     string `json:"author,omitempty"`
	Context    string `json:"context,omitempty"`
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
}

type Author struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type FeaturedImage struct {
	URL     string `json:"url"`
	Caption string `json:"caption,omitempty"`
	Credit  string `json:"credit,omitempty"`
}

type SocialVariants struct {
	X          string `json:"x,omitempty"`
	Facebook   string `json:"facebook,omitempty"`
	LinkedIn   string `json:"linkedin,omitempty"`
	Newsletter string `json:"newsletter,omitempty"`
}

type SourceLink struct {
	URL        string `json:"url"`
	Publisher  string `json:"publisher"`
	AccessedAt string `json:"accessedAt"`
}

type Article struct {
	ID             string           `json:"id"`
	Slug           string           `json:"slug"`
	Title          string           `json:"title"`
	Kicker         string           `json:"kicker"`
	Deck           string           `json:"deck"`
	Category       PillarCategory   `json:"category"`
	CategoryName   string           `json:"categoryName"`
	Author         Author           `json:"author"`
	FeaturedImage  FeaturedImage    `json:"featuredImage"`
	ReadTime       string           `json:"readTime"`
	Status         PublishingStatus `json:"status"`
	Body           []ContentBlock   `json:"body"`
	SourceLinks    []SourceLink     `json:"sourceLinks"`
	SocialVariants SocialVariants   `json:"socialVariants"`
	PublishedAt    time.Time        `json:"publishedAt"`
	CreatedAt      time.Time        `json:"createdAt"`
}
