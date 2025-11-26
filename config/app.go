package config

import (
	"be_uas/app/repository/postgres"
	"be_uas/app/service"
	"be_uas/database"
	"be_uas/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func NewApp() *fiber.App {
	app := fiber.New()

	// Register Middleware Global
	app.Use(cors.New())
	app.Use(logger.New(NewLoggerConfig())) // logger custom

	// Dependency Injection
	// Init Repository
	userRepo := postgres.NewUserRepo(database.DB)

	// Init Service
	authService := service.NewAuthService(userRepo)

	// Register Routes
	route.SetupRoutes(app, authService)

	return app
}