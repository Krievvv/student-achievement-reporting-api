package main

import (
	"be_uas/config"
	"be_uas/database"
	"log"
	"os"
)

func main() {
	// Load Env
	config.LoadEnv()

	// Connect Databases
	database.ConnectPostgres()
	database.ConnectMongo() 

	// Setup App
	app := config.NewApp()

	// Start Server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(":" + port))
}