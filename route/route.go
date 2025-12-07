package route

import (
	"be_uas/app/service"
	"be_uas/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, 
	authService *service.AuthService, 
	achieveService *service.AchievementService, 
	adminService *service.AdminService, 
	reportService *service.ReportService,
	academicService *service.AcademicService) { 

	api := app.Group("/api/v1")

	// AUTHENTICATION
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/seed", authService.SeedAdmin)
	
	// Endpoint Auth
	auth.Post("/refresh", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"token": "new_refreshed_token"}) }) // Mockup
	auth.Post("/logout", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"message": "Logged out"}) }) // Mockup
	auth.Get("/profile", middleware.AuthRequired(), func(c *fiber.Ctx) error { 
		return c.JSON(fiber.Map{"user_id": c.Locals("user_id"), "role": c.Locals("role")}) 
	})

	// Admin
	admin := api.Group("/users", middleware.AuthRequired(), middleware.PermissionCheck("Admin")) 
	admin.Post("/", adminService.CreateUser)
	admin.Get("/", adminService.ListUsers)
	admin.Delete("/:id", adminService.DeleteUser)
	
	// Tambahan Endpoint User
	admin.Get("/:id", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"data": "User Detail"}) }) 
	admin.Put("/:id", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"message": "User Updated"}) }) 
	admin.Put("/:id/role", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"message": "Role Updated"}) })

	// ACHIEVEMENTS
	ach := api.Group("/achievements", middleware.AuthRequired())

	// Endpoint Umum (Milik Mahasiswa Sendiri / Detail)
	ach.Get("/", achieveService.GetMyAchievements) 
	ach.Get("/:id", achieveService.GetAchievementDetail) 
	ach.Post("/:id/attachments", achieveService.UploadAttachment) 
	ach.Get("/:id/history", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"history": []string{"draft"}}) }) // Mockup
	
	// Endpoint Action Mahasiswa
	ach.Post("/", middleware.PermissionCheck("Mahasiswa"), achieveService.CreateAchievement)
	ach.Put("/:id", middleware.PermissionCheck("Mahasiswa"), func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"message": "Updated"}) }) // Mockup Update
	ach.Post("/:id/submit", middleware.PermissionCheck("Mahasiswa"), achieveService.SubmitForVerification)
	ach.Delete("/:id", middleware.PermissionCheck("Mahasiswa"), achieveService.DeleteAchievement)

	// Endpoint Action Dosen Wali
	ach.Get("/advisees", middleware.PermissionCheck("Dosen Wali"), achieveService.GetAdviseeAchievements)
	ach.Post("/:id/verify", middleware.PermissionCheck("Dosen Wali"), achieveService.VerifyAchievement)
	ach.Post("/:id/reject", middleware.PermissionCheck("Dosen Wali"), achieveService.RejectAchievement)

	// Global View (Admin)
	ach.Get("/all", middleware.PermissionCheck("Admin"), adminService.GetAllAchievements)

	// STUDENTS & LECTURERS 
	std := api.Group("/students", middleware.AuthRequired())
	std.Get("/", academicService.GetAllStudents)
	std.Get("/:id", academicService.GetStudentByID)
	std.Put("/:id/advisor", middleware.PermissionCheck("Admin"), academicService.UpdateStudentAdvisor)
	std.Get("/:id/achievements", achieveService.GetStudentAchievements) 

	lec := api.Group("/lecturers", middleware.AuthRequired())
	lec.Get("/", academicService.GetAllLecturers)
	lec.Get("/:id/advisees", academicService.GetLecturerAdvisees)

	// 5.8 REPORTS
	reports := api.Group("/reports", middleware.AuthRequired())
	reports.Get("/statistics", reportService.GetStatistics)
	reports.Get("/student/:id", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"report": "Student Report"}) })
}