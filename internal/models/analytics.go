package models

import (
	"time"
)

// Basic Analytics Models (simplified for integration testing)

// Recommendation represents basic actionable insights
type Recommendation struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`     // scaling, optimization, configuration
	Priority    string    `json:"priority"` // high, medium, low
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	Effort      string    `json:"effort"` // low, medium, high
	CreatedAt   time.Time `json:"created_at"`
}
