package models

import "time"

// PostSubscriber represents an email address registered for article publishing notifications.
type PostSubscriber struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	Categories []string  `json:"categories"` // empty array means all categories
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// EmailDispatch represents a notification record or queue entry for a published article.
type EmailDispatch struct {
	ID             string     `json:"id"`
	ArticleID      string     `json:"articleId"`
	RecipientEmail string     `json:"recipientEmail"`
	Status         string     `json:"status"` // 'pending', 'queued', 'sent', 'failed'
	ErrorMessage   string     `json:"errorMessage,omitempty"`
	SentAt         *time.Time `json:"sentAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}
