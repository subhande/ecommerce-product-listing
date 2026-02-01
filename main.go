package main

import (
	"ecommerce_product_listing/config"
	"ecommerce_product_listing/handler"
	"ecommerce_product_listing/repository"
	"ecommerce_product_listing/service"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	config.LoadEnv()
	config.ConnectDB()
	config.Initialize()

	repo := &repository.ProductRepository{}
	service := &service.ProductService{Repo: repo}
	handler := &handler.ProductHandler{Service: service}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Default 500
			code := fiber.StatusInternalServerError

			// If it's a fiber error, get status code
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			// Print actual error in console
			fmt.Println("ERROR:", err)

			// Send response
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${error}\n",
	}))

	api := app.Group("/api")
	v1 := api.Group("/v1")

	products := v1.Group("/products")

	products.Get("/", handler.GetProducts)
	products.Post("/", handler.AddProduct)
	products.Post("/bulk", handler.AddProductsBulk)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to the E-commerce Product Listing API",
		})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Listen(":8080")
}
