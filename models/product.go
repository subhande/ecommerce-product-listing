package models

import "time"

type Product struct {
	ID                int        `json:"id,omitempty"`
	Title             string     `json:"title"`
	ASIN              string     `json:"asin,omitempty"`
	Description       string     `json:"description,omitempty"`
	Category          string     `json:"category,omitempty"`
	Brand             string     `json:"brand,omitempty"`
	ImageURL          string     `json:"image_url,omitempty"`
	ProductURL        string     `json:"product_url,omitempty"`
	Price             float64    `json:"price"`
	Currency          string     `json:"currency"`
	Country           string     `json:"country,omitempty"`
	Stock             int        `json:"stock"`
	AvgRating         float64    `json:"avg_rating,omitempty"`
	ReviewCount       int        `json:"review_count,omitempty"`
	BoughtInLastMonth int        `json:"bought_in_last_month,omitempty"`
	IsBestSeller      bool       `json:"is_best_seller,omitempty"`
	CreatedAt         *time.Time `json:"created_at,omitempty"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty"`
}
