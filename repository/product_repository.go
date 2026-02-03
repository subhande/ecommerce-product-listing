package repository

import (
	"context"
	"ecommerce_product_listing/config"
	"ecommerce_product_listing/models"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type ProductRepository struct{}

func (r *ProductRepository) CreateProduct(
	ctx context.Context,
	p *models.Product,
) (*models.Product, error) {

	query := `
	INSERT INTO products (title, asin, description, category, brand, image_url, product_url, price, currency, country, stock, avg_rating, review_count, bought_in_last_month, is_best_seller, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW())
	RETURNING id, created_at, updated_at
	`

	err := config.DB.QueryRow(
		ctx,
		query,
		p.Title,
		p.ASIN,
		p.Description,
		p.Category,
		p.Brand,
		p.ImageURL,
		p.ProductURL,
		p.Price,
		p.Currency,
		p.Country,
		p.Stock,
		p.AvgRating,
		p.ReviewCount,
		p.BoughtInLastMonth,
		p.IsBestSeller,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *ProductRepository) CreateProductsBulk(
	ctx context.Context,
	products []models.Product,
) ([]models.Product, error) {

	tx, err := config.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	query := `
	INSERT INTO products (title, asin, description, category, brand, image_url, product_url, price, currency, country, stock, avg_rating, review_count, bought_in_last_month, is_best_seller, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW())
	RETURNING id, created_at, updated_at
	`

	for _, p := range products {
		batch.Queue(query,
			p.Title,
			p.ASIN,
			p.Description,
			p.Category,
			p.Brand,
			p.ImageURL,
			p.ProductURL,
			p.Price,
			p.Currency,
			p.Country,
			p.Stock,
			p.AvgRating,
			p.ReviewCount,
			p.BoughtInLastMonth,
			p.IsBestSeller,
		)
	}

	br := tx.SendBatch(ctx, batch)

	for i := range products {
		err := br.QueryRow().Scan(
			&products[i].ID,
			&products[i].CreatedAt,
			&products[i].UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
	}

	err = br.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (r *ProductRepository) GetProducts(
	ctx context.Context,
	category string,
	minPrice float64,
	maxPrice float64,
	limit int,
	offset int,
) ([]models.Product, error) {

	query := `SELECT id, name, description, price, stock, category, created_at, updated_at
			  FROM products WHERE 1=1`

	args := []interface{}{}
	argPos := 1

	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", argPos)
		args = append(args, category)
		argPos++
	}

	if minPrice > 0 {
		query += fmt.Sprintf(" AND price >= $%d", argPos)
		args = append(args, minPrice)
		argPos++
	}

	if maxPrice > 0 {
		query += fmt.Sprintf(" AND price <= $%d", argPos)
		args = append(args, maxPrice)
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := config.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []models.Product{}

	for rows.Next() {
		var p models.Product
		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Description,
			&p.Price,
			&p.Stock,
			&p.Category,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}
