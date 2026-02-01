package service

import (
	"context"
	"ecommerce_product_listing/models"
	"ecommerce_product_listing/repository"
)

type ProductService struct {
	Repo *repository.ProductRepository
}

func (s *ProductService) AddProduct(
	ctx context.Context,
	p *models.Product,
) (*models.Product, error) {

	return s.Repo.CreateProduct(ctx, p)
}

func (s *ProductService) AddProductsBulk(
	ctx context.Context,
	products []models.Product,
) ([]models.Product, error) {

	return s.Repo.CreateProductsBulk(ctx, products)
}

func (s *ProductService) ListProducts(
	ctx context.Context,
	category string,
	minPrice float64,
	maxPrice float64,
	limit int,
	offset int,
) ([]models.Product, error) {

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	return s.Repo.GetProducts(ctx, category, minPrice, maxPrice, limit, offset)
}
