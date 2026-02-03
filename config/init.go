package config

import (
	"context"
	"log"
)

func Initialize() {

	// // Create database if not exists
	// DATABASE_NAME := os.Getenv("DATABASE_NAME")
	// _, err := DB.Exec(context.Background(), "CREATE DATABASE "+DATABASE_NAME)
	// if err != nil {
	// 	log.Println("Database already exists or error creating database:", err)
	// } else {
	// 	log.Println("Database created successfully")
	// }

	// // Switch to the created database
	// _, err = DB.Exec(context.Background(), "USE "+DATABASE_NAME)
	// if err != nil {
	// 	log.Println("Error switching to database:", err)
	// } else {
	// 	log.Println("Switched to database:", DATABASE_NAME)
	// }

	// Create products table if not exists
	_, err := DB.Exec(context.Background(), `
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
		log.Println("Products table created or already exists")
	}

	// Create Index on category for faster queries
	_, err = DB.Exec(context.Background(), `
	CREATE INDEX IF NOT EXISTS idx_category ON products(category)`)
	if err != nil {
		log.Println("Error creating index on category:", err)
	} else {
		log.Println("Index on category created or already exists")
	}

	// Create Index on price for faster queries
	_, err = DB.Exec(context.Background(), `
	CREATE INDEX IF NOT EXISTS idx_price ON products(price)`)
	if err != nil {
		log.Println("Error creating index on price:", err)
	} else {
		log.Println("Index on price created or already exists")
	}

}
