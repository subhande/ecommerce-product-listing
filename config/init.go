package config

import (
	"context"
	"log"
)

func Initialize() {
	var err error

	// Drop table if exists (for testing purposes)
	_, err = DB.Exec(context.Background(), `DROP TABLE IF EXISTS products`)
	if err != nil {
		log.Println("Error dropping products table:", err)
	} else {
		log.Println("Products table dropped if it existed")
	}

	// Check if products table exists, if not create it and add indexes
	var exists bool
	err = DB.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'products'
		)`).Scan(&exists)
	if err != nil {
		log.Println("Error checking products table existence:", err)
		return
	}

	if exists {
		log.Println("Products table already exists")
	} else {
		log.Println("Products table does not exist, creating...")

		// Create products table if not exists
		_, err = DB.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS products (
			id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			title TEXT NOT NULL,
			asin VARCHAR(255) UNIQUE,
			description TEXT,
			category VARCHAR(255),
			brand VARCHAR(255),
			image_url TEXT,
			product_url TEXT,
			price NUMERIC(10, 2) NOT NULL,
			currency VARCHAR(10) NOT NULL,
			country VARCHAR(50),
			stock INT NOT NULL DEFAULT 0,
			avg_rating NUMERIC(3, 2),
			review_count INT,
			bought_in_last_month INT,
			is_best_seller BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`)
		if err != nil {
			log.Println("Error creating products table:", err)
		} else {
			log.Println("Products table created")
		}
	}

	queries := []string{
		// Category Indexes for faster queries
		"CREATE INDEX idx_products_popular_keyset ON products (bought_in_last_month DESC, id DESC) WHERE stock > 0;",
		"CREATE INDEX idx_products_price_asc_keyset ON products (price ASC, id ASC) WHERE stock > 0;",
		"CREATE INDEX idx_products_rating_keyset ON products (avg_rating DESC, id DESC) WHERE stock > 0;",
		"CREATE INDEX idx_products_updated_keyset ON products (updated_at DESC, id DESC) WHERE stock > 0;",
		// Composite Indexes for category + sorting
		"CREATE INDEX idx_products_cat_popular ON products (category, bought_in_last_month DESC, id DESC) WHERE stock > 0;",
		"CREATE INDEX idx_products_cat_price ON products (category, price ASC, id ASC) WHERE stock > 0;",
		"CREATE INDEX idx_products_cat_rating ON products (category, avg_rating DESC, id DESC) WHERE stock > 0;",
		"CREATE INDEX idx_products_cat_updated ON products (category, updated_at DESC, id DESC) WHERE stock > 0;",
		// Composite Indexes for brand + sorting
		"CREATE INDEX idx_products_brand_popular ON products (brand, bought_in_last_month DESC, id DESC) WHERE stock > 0;",
		"CREATE INDEX idx_products_brand_price ON products (brand, price ASC, id ASC) WHERE stock > 0;",
		"CREATE INDEX idx_products_brand_rating ON products (brand, avg_rating DESC, id DESC) WHERE stock > 0;",
		"CREATE INDEX idx_products_brand_updated ON products (brand, updated_at DESC, id DESC) WHERE stock > 0;",
		// Full Text Search Indexes for title and description
		"CREATE EXTENSION IF NOT EXISTS pg_trgm;",
		"CREATE INDEX idx_products_search_trgm ON products USING gin (title gin_trgm_ops, description gin_trgm_ops);",
	}

	for _, query := range queries {
		_, err := DB.Exec(context.Background(), query)
		if err != nil {
			log.Println("Error executing query:", err)
		} else {
			log.Println("Executed query successfully")
		}
	}

}
