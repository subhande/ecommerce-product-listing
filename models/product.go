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

type SortByEnum string
type SortOrderEnum string

const (
	SortByPrice            SortByEnum = "price"
	SortByPopularity       SortByEnum = "bought_in_last_month"
	SortByRating           SortByEnum = "avg_rating"
	SortByModificationDate SortByEnum = "updated_at"
)

const (
	SortOrderAsc  SortOrderEnum = "asc"
	SortOrderDesc SortOrderEnum = "desc"
)

func (s SortByEnum) IsValid() bool {
	switch s {
	case SortByPrice, SortByPopularity, SortByRating, SortByModificationDate:
		return true
	}
	return false
}

func (o SortOrderEnum) IsValid() bool {
	return o == SortOrderAsc || o == SortOrderDesc
}

type ProductFilter struct {
	SearchQueryText     string        `query:"search_query_text,omitempty"`
	Category            string        `query:"category,omitempty"`
	Brand               string        `query:"brand,omitempty"`
	MinPrice            float64       `query:"min_price,omitempty"`
	MaxPrice            float64       `query:"max_price,omitempty"`
	ShowOutOfStock      bool          `query:"show_out_of_stock,omitempty"`
	RatingMoreThanEqual float64       `query:"rating_more_than_equal,omitempty"`
	ReviewCount         int           `query:"review_count,omitempty"`
	SortByColumn        SortByEnum    `query:"sort_by_column,omitempty"`
	SortOrder           SortOrderEnum `query:"sort_order,omitempty"`
	SortLastValue       string        `query:"sort_last_value,omitempty"`
	LastID              int           `query:"last_id,omitempty"`
	PageSize            int           `query:"page_size,omitempty"`
}

func (f *ProductFilter) Normalize() {
	if f.PageSize <= 0 || f.PageSize > 100 {
		f.PageSize = 20
	}

	if f.MinPrice != -1 && f.MaxPrice != -1 && f.MinPrice > f.MaxPrice {
		f.MinPrice, f.MaxPrice = -1, -1
	}

	if !f.SortByColumn.IsValid() {
		f.SortByColumn = SortByPopularity
	}
	if !f.SortOrder.IsValid() {
		f.SortOrder = SortOrderDesc
	}
}

const (
	DefaultPageSize            = 20
	MaxPageSize                = 100
	DefaultMinPrice            = -1
	DefaultMaxPrice            = -1
	DefaultSortBy              = SortByPopularity
	DefaultSortOrder           = SortOrderDesc
	DefaultLastID              = -1
	DefaultReviewCount         = -1
	DefaultShowOutOfStock      = false
	DefaultRatingMoreThanEqual = -1
)

func NewProductFilter() *ProductFilter {
	return &ProductFilter{
		PageSize:            DefaultPageSize,
		MinPrice:            DefaultMinPrice,
		MaxPrice:            DefaultMaxPrice,
		SortByColumn:        DefaultSortBy,
		SortOrder:           DefaultSortOrder,
		LastID:              DefaultLastID,
		ReviewCount:         DefaultReviewCount,
		ShowOutOfStock:      DefaultShowOutOfStock,
		RatingMoreThanEqual: DefaultRatingMoreThanEqual,
	}
}
