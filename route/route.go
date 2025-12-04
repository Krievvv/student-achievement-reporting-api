package route

import (
	"be_uas/app/service"
	"be_uas/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authService *service.AuthService, achieveService *service.AchievementService, adminService *service.AdminService, reportService *service.ReportService) {
	api := app.Group("/api/v1")

	// Auth
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/seed", authService.SeedAdmin)

	// Achievement (Protected) 
	ach := api.Group("/achievements", middleware.AuthRequired())
	
	// Endpoint Mahasiswa
	ach.Post("/", middleware.PermissionCheck("Mahasiswa"), achieveService.CreateAchievement)
	ach.Post("/:id/submit", middleware.PermissionCheck("Mahasiswa"), achieveService.SubmitForVerification)
	ach.Delete("/:id", middleware.PermissionCheck("Mahasiswa"), achieveService.DeleteAchievement)

	// Endpoint Dosen Wali
	ach.Get("/advisees", middleware.PermissionCheck("Dosen Wali"), achieveService.GetAdviseeAchievements) // View List
	ach.Post("/:id/verify", middleware.PermissionCheck("Dosen Wali"), achieveService.VerifyAchievement)
	ach.Post("/:id/reject", middleware.PermissionCheck("Dosen Wali"), achieveService.RejectAchievement)

	// ADMIN ROUTES
	admin := api.Group("/admin", middleware.AuthRequired(), middleware.PermissionCheck("Admin"))
	
	// Manage Users
	admin.Post("/users", adminService.CreateUser)
	admin.Get("/users", adminService.ListUsers)
	admin.Delete("/users/:id", adminService.DeleteUser)

	// Global View Achievements (Pagination)
	admin.Get("/achievements", adminService.GetAllAchievements)

	// REPORT ROUTES
	reports := api.Group("/reports", middleware.AuthRequired())
	reports.Get("/statistics", reportService.GetStatistics)
}