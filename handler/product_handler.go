package handler

import (
	"context"
	"ecommerce_product_listing/models"
	"ecommerce_product_listing/service"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type ProductHandler struct {
	Service *service.ProductService
}

func (h *ProductHandler) AddProduct(c *fiber.Ctx) error {

	var product models.Product

	if err := c.BodyParser(&product); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if product.Title == "" || product.Price <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name and price are required",
		})
	}

	now := time.Now()

	if product.CreatedAt == nil {
		product.CreatedAt = &now
	}

	if product.UpdatedAt == nil {
		product.UpdatedAt = &now
	}

	result, err := h.Service.AddProduct(context.Background(), &product)
	if err != nil {
		log.Error("Failed to create product:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create product",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *ProductHandler) AddProductsBulk(c *fiber.Ctx) error {

	var products []models.Product

	if err := c.BodyParser(&products); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if len(products) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "empty product list",
		})
	}

	now := time.Now()

	for i := range products {
		if products[i].CreatedAt == nil {
			products[i].CreatedAt = &now
		}

		if products[i].UpdatedAt == nil {
			products[i].UpdatedAt = &now
		}
	}

	result, err := h.Service.AddProductsBulk(context.Background(), products)
	if err != nil {
		log.Error("Failed to insert products:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to insert products",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {

	productFilter := models.NewProductFilter()

	// productFilter := &models.ProductFilter{}

	// Log incoming query parameters
	log.Info("Incoming query parameters:", fmt.Sprintf("%v", c.Queries()))

	if err := c.QueryParser(productFilter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":         "invalid request body",
			"error_message": err.Error(),
		})
	}
	log.Info("Parsed product filter:", fmt.Sprintf("%+v", productFilter))
	products, err := h.Service.ListProducts(
		c.Context(), // fasthttp context
		productFilter,
	)

	if err != nil {
		log.Error("Failed to fetch products:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch products",
		})
	}

	count := len(products)

	var lastID int
	var sortLastValue interface{}

	if count > 0 {
		lastID = products[count-1].ID
		if string(productFilter.SortByColumn) != "" {
			product := products[count-1]
			switch productFilter.SortByColumn {
			case models.SortByPrice:
				sortLastValue = fmt.Sprintf("%f", product.Price)
			case models.SortByPopularity:
				sortLastValue = fmt.Sprintf("%d", product.BoughtInLastMonth)
			case models.SortByRating:
				sortLastValue = fmt.Sprintf("%f", product.AvgRating)
			case models.SortByModificationDate:
				sortLastValue = product.UpdatedAt.Format(time.RFC3339Nano)
			default:
				sortLastValue = nil
			}
		}
	}

	return c.JSON(fiber.Map{
		"count":           len(products),
		"last_id":         lastID,
		"sort_order":      productFilter.SortOrder,
		"sort_last_value": sortLastValue,
		"sort_by_column":  productFilter.SortByColumn,
		"products":        products,
	})
}
