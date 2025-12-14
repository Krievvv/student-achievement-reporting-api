package route

import (
	"be_uas/app/service"
    "be_uas/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRoutes(app *fiber.App, 
	authS *service.AuthService, 
	adminS *service.AdminService, 
	achS *service.AchievementService, 
	repS *service.ReportService,
	acadS *service.AcademicService) {
	
	app.Use(logger.New())

	api := app.Group("/api/v1")

	AuthRoutes(api, authS)
	UserRoutes(api, adminS) 
	AchievementRoutes(api, achS)
	AcademicRoutes(api, acadS, achS)

	api.Get("/reports/statistics", middleware.AuthRequired(), repS.GetStatistics)
	api.Get("/reports/student/:id", middleware.AuthRequired(), repS.GetStudentReport)
}