package route

import (
	"be_uas/app/service"
	"be_uas/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authService *service.AuthService) {
	api := app.Group("/api/v1")

	// Auth
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/seed", authService.SeedAdmin)

	// Test Token Route
	api.Get("/check-token", middleware.AuthRequired(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Access Granted",
			"user":    c.Locals("username"),
		})
	})
}