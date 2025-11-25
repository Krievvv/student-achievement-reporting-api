package route

import (
	"be_uas/app/service"
	// Import middleware dari lokasi baru
	"be_uas/middleware" 

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authService *service.AuthService) {
	api := app.Group("/api/v1")

	// Auth Routes (Public)
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/seed", authService.SeedAdmin)

	// Contoh Proteksi Route (Test Token)
	// Hanya bisa diakses jika punya token
	api.Get("/check-token", middleware.AuthRequired(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Token Valid!",
			"user_id": c.Locals("user_id"),
			"role":    c.Locals("role"),
		})
	})
}