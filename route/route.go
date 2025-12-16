package route

import (
	"be_uas/app/service"
    "be_uas/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger" 
	_ "be_uas/docs"
)

func SetupRoutes(app *fiber.App, 
	authS *service.AuthService, 
	adminS *service.AdminService, 
	achS *service.AchievementService, 
	repS *service.ReportService,
	acadS *service.AcademicService) {
	
	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
        return c.Redirect("/swagger/index.html")
    })
	app.Get("/swagger/*", swagger.HandlerDefault)

	api := app.Group("/api/v1")

	AuthRoutes(api, authS)
	UserRoutes(api, adminS) 
	AchievementRoutes(api, achS)
	AcademicRoutes(api, acadS, achS)

	api.Get("/reports/statistics", middleware.AuthRequired(), repS.GetStatistics)
	api.Get("/reports/student/:id", middleware.AuthRequired(), repS.GetStudentReport)
}