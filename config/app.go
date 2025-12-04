package config

import (
	// Repository
	repoMongo "be_uas/app/repository/mongodb"
	repoPG "be_uas/app/repository/postgres"
	
	// Service
	"be_uas/app/service"
	
	// DB Drivers
	"be_uas/database"
	"be_uas/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func NewApp() *fiber.App {
	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New(NewLoggerConfig()))

	// DEPENDENCY INJECTION

	// Auth Module
	userRepo := repoPG.NewUserRepo(database.DB)
	authService := service.NewAuthService(userRepo)

	// Achievement Module
	achieveRepoPG := repoPG.NewAchievementRepoPG(database.DB)
	achieveRepoMongo := repoMongo.NewAchievementRepoMongo(database.MongoDB) // Gunakan var MongoDB
	achieveService := service.NewAchievementService(achieveRepoPG, achieveRepoMongo)
	adminService := service.NewAdminService(userRepo, achieveRepoPG, achieveRepoMongo)

	// Report Service
	reportService := service.NewReportService(reportRepoPG, reportRepoMongo)

	// Report Repos
	reportRepoPG := repoPG.NewReportRepoPG(database.DB)
	reportRepoMongo := repoMongo.NewReportRepoMongo(database.MongoDB)

	// ROUTES
	route.SetupRoutes(app, authService, achieveService, adminService, reportService)

	return app
}