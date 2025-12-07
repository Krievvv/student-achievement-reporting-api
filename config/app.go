package config

import (
	// Repository
	repoMongo "be_uas/app/repository/mongodb"
	repoPG "be_uas/app/repository/postgres" 

	// Service
	"be_uas/app/service"

	// DB Drivers & Routes
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

	// ==================================
	// DEPENDENCY INJECTION
	// ==================================

	// 1. Init Repositories (Postgres & Mongo)
	userRepo := repoPG.NewUserRepo(database.DB)
	achieveRepoPG := repoPG.NewAchievementRepoPG(database.DB)
	achieveRepoMongo := repoMongo.NewAchievementRepoMongo(database.MongoDB)
	
	reportRepoPG := repoPG.NewReportRepoPG(database.DB)
	reportRepoMongo := repoMongo.NewReportRepoMongo(database.MongoDB)

	// ==> PERBAIKAN: Init Academic Repo DI SINI (Sebelum Service)
	academicRepo := repoPG.NewAcademicRepoPG(database.DB) 

	// 2. Init Services
	authService := service.NewAuthService(userRepo)
	achieveService := service.NewAchievementService(achieveRepoPG, achieveRepoMongo)
	adminService := service.NewAdminService(userRepo, achieveRepoPG, achieveRepoMongo)
	reportService := service.NewReportService(reportRepoPG, reportRepoMongo)
	
	// ==> PERBAIKAN: Init Academic Service DI SINI (Sebelum SetupRoutes)
	academicService := service.NewAcademicService(academicRepo)

	// ==================================
	// ROUTES
	// ==================================
	
	// ==> PERBAIKAN: Masukkan academicService sebagai argumen terakhir
	route.SetupRoutes(app, authService, achieveService, adminService, reportService, academicService)

	return app
}