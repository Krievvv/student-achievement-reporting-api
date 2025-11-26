package route

import (
	"be_uas/app/service"
	"be_uas/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authService *service.AuthService, achieveService *service.AchievementService) {
	api := app.Group("/api/v1")

	// Auth
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/seed", authService.SeedAdmin)

	// Achievement (Protected) 
	ach := api.Group("/achievements", middleware.AuthRequired())
	
	// Create
	ach.Post("/", middleware.PermissionCheck("Mahasiswa"), achieveService.CreateAchievement)
	
	// Submit
	ach.Post("/:id/submit", middleware.PermissionCheck("Mahasiswa"), achieveService.SubmitForVerification)
	
	// Delete
	ach.Delete("/:id", middleware.PermissionCheck("Mahasiswa"), achieveService.DeleteAchievement)
}