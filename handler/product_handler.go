package handler

import (
	"context"
	"ecommerce_product_listing/models"
	"ecommerce_product_listing/service"
	"strconv"
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

	category := c.Query("category")

	minPrice, _ := strconv.ParseFloat(c.Query("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("max_price"), 64)

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	products, err := h.Service.ListProducts(
		c.Context(), // fasthttp context
		category,
		minPrice,
		maxPrice,
		limit,
		offset,
	)

	if err != nil {
		log.Error("Failed to fetch products:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch products",
		})
	}

	return c.JSON(fiber.Map{
		"count": len(products),
		"data":  products,
	})
}
