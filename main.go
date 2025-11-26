package main

import (
	"be_uas/config"
	"be_uas/database"
	"log"
	"os"
)

func main() {
	// Load Environment Variables
	config.LoadEnv()

	// Connect to Database
	database.ConnectPostgres()

	// Setup App (Fiber, Middleware, Routes, DI)
	app := config.NewApp()

	// Start Server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(":" + port))
}