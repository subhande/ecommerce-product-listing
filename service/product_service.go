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
	prodcutFilter *models.ProductFilter,
) ([]models.Product, error) {

	return s.Repo.GetProducts(ctx, prodcutFilter)
}
