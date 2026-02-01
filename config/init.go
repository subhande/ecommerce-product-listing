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
		name VARCHAR(100) NOT NULL,
		description TEXT,
		price NUMERIC(10, 2) NOT NULL,
		stock INT NOT NULL DEFAULT 0,
		category VARCHAR(50),
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
