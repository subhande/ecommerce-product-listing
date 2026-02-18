package repository

import (
	"context"
	"ecommerce_product_listing/config"
	"ecommerce_product_listing/models"
	"fmt"
	"log"
	"strconv"
	"time"

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
	productFilter *models.ProductFilter,
) ([]models.Product, error) {

	productFilter.Normalize()

	query := `SELECT 
	id, title, asin, description, category, brand, image_url, product_url, price, currency, country, stock, avg_rating, review_count, bought_in_last_month, is_best_seller, created_at, updated_at
	FROM products WHERE 1=1 `

	args := []interface{}{}
	argPos := 1

	if productFilter.SearchQueryText != "" && productFilter.SearchType == models.SimpleTextSearchType {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argPos, argPos)
		searchPattern := "%" + productFilter.SearchQueryText + "%"
		args = append(args, searchPattern)
		argPos++
	}

	if productFilter.SearchQueryText != "" && productFilter.SearchType == models.VectorSearchType {
		query += fmt.Sprintf(" AND search_vector @@ websearch_to_tsquery('english', $%d)", argPos)
		args = append(args, productFilter.SearchQueryText)
		argPos++
	}

	if productFilter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argPos)
		args = append(args, productFilter.Category)
		argPos++
	}

	if productFilter.Brand != "" {
		query += fmt.Sprintf(" AND brand = $%d", argPos)
		args = append(args, productFilter.Brand)
		argPos++
	}

	if productFilter.MinPrice != -1 {
		query += fmt.Sprintf(" AND price >= $%d", argPos)
		args = append(args, productFilter.MinPrice)
		argPos++
	}

	if productFilter.MaxPrice != -1 {
		query += fmt.Sprintf(" AND price <= $%d", argPos)
		args = append(args, productFilter.MaxPrice)
		argPos++
	}

	if !productFilter.ShowOutOfStock {
		query += " AND stock > 0"
	}

	if productFilter.RatingMoreThanEqual > 0 {
		query += fmt.Sprintf(" AND avg_rating >= $%d", argPos)
		args = append(args, productFilter.RatingMoreThanEqual)
		argPos++
	}

	if productFilter.ReviewCount > 0 {
		query += fmt.Sprintf(" AND review_count >= $%d", argPos)
		args = append(args, productFilter.ReviewCount)
		argPos++
	}

	if productFilter.LastID != -1 && productFilter.SortLastValue != "" && productFilter.PageNumber == -1 {
		operator := ">"
		if productFilter.SortOrder == models.SortOrderDesc {
			operator = "<"
		}

		var sortLastValue interface{}
		var err error
		switch productFilter.SortByColumn {
		case models.SortByPrice, models.SortByRating:
			sortLastValue, err = strconv.ParseFloat(productFilter.SortLastValue, 64)
		case models.SortByPopularity:
			sortLastValue, err = strconv.Atoi(productFilter.SortLastValue)
		case models.SortByModificationDate:
			sortLastValue, err = time.Parse(time.RFC3339Nano, productFilter.SortLastValue)
		default:
			err = fmt.Errorf("unsupported sort column: %s", productFilter.SortByColumn)
		}
		if err != nil {
			return nil, fmt.Errorf("invalid cursor value for %s: %w", productFilter.SortByColumn, err)
		}
		query += fmt.Sprintf(" AND (%s, id) %s ($%d, $%d)", productFilter.SortByColumn, operator, argPos, argPos+1)
		args = append(args, sortLastValue, productFilter.LastID)
		argPos += 2
	}

	// Order By KeySet (SortByColumn, ID)
	direction := "ASC"
	if productFilter.SortOrder == models.SortOrderDesc {
		direction = "DESC"
	}
	if productFilter.PageNumber > 0 {
		query += fmt.Sprintf(" ORDER BY %s %s", productFilter.SortByColumn, direction)
	} else {
		query += fmt.Sprintf(" ORDER BY %s %s, id %s", productFilter.SortByColumn, direction, direction)
	}
	query += fmt.Sprintf(" LIMIT $%d", argPos)
	args = append(args, productFilter.PageSize)
	argPos++

	if productFilter.PageNumber > 0 {
		offset := (productFilter.PageNumber - 1) * productFilter.PageSize
		if offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argPos)
			args = append(args, offset)
			argPos++
		}
	}

	log.Printf("Constructed SQL query: %s", query)
	log.Printf("With arguments: %v", args)

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
			&p.ASIN,
			&p.Description,
			&p.Category,
			&p.Brand,
			&p.ImageURL,
			&p.ProductURL,
			&p.Price,
			&p.Currency,
			&p.Country,
			&p.Stock,
			&p.AvgRating,
			&p.ReviewCount,
			&p.BoughtInLastMonth,
			&p.IsBestSeller,
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

func (r *ProductRepository) GetCounts(
	ctx context.Context,
	productFilter *models.ProductFilter,
) (int64, error) {
	productFilter.Normalize()

	query := `SELECT COUNT(*) FROM products WHERE 1=1 `
	args := []interface{}{}
	argPos := 1

	if productFilter.SearchQueryText != "" && productFilter.SearchType == models.SimpleTextSearchType {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argPos, argPos)
		searchPattern := "%" + productFilter.SearchQueryText + "%"
		args = append(args, searchPattern)
		argPos++
	}

	if productFilter.SearchQueryText != "" && productFilter.SearchType == models.VectorSearchType {
		query += fmt.Sprintf(" AND search_vector @@ websearch_to_tsquery('english', $%d)", argPos)
		args = append(args, productFilter.SearchQueryText)
		argPos++
	}

	if productFilter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argPos)
		args = append(args, productFilter.Category)
		argPos++
	}

	if productFilter.Brand != "" {
		query += fmt.Sprintf(" AND brand = $%d", argPos)
		args = append(args, productFilter.Brand)
		argPos++
	}

	if productFilter.MinPrice != -1 {
		query += fmt.Sprintf(" AND price >= $%d", argPos)
		args = append(args, productFilter.MinPrice)
		argPos++
	}

	if productFilter.MaxPrice != -1 {
		query += fmt.Sprintf(" AND price <= $%d", argPos)
		args = append(args, productFilter.MaxPrice)
		argPos++
	}

	if !productFilter.ShowOutOfStock {
		query += " AND stock > 0"
	}

	if productFilter.RatingMoreThanEqual > 0 {
		query += fmt.Sprintf(" AND avg_rating >= $%d", argPos)
		args = append(args, productFilter.RatingMoreThanEqual)
		argPos++
	}

	if productFilter.ReviewCount > 0 {
		query += fmt.Sprintf(" AND review_count >= $%d", argPos)
		args = append(args, productFilter.ReviewCount)
		argPos++
	}

	log.Printf("Constructed SQL count query: %s", query)
	log.Printf("With arguments: %v", args)

	var count int64
	err := config.DB.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
