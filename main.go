package main

import (
	"be_uas/app/repository/postgres"
	"be_uas/app/service"
	"be_uas/config"
	"be_uas/route"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 2. Connect DB (Pastikan .env sudah sslmode=disable)
	config.ConnectPostgres()
	
	// 3. Dependency Injection
	userRepo := postgres.NewUserRepo(config.DB)
	authService := service.NewAuthService(userRepo)

	// 4. Fiber App
	app := fiber.New()
	app.Use(logger.New())

	// 5. Routes
	route.SetupRoutes(app, authService)

	// 6. Start Server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(":" + port))
}