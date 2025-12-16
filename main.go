package main

import (
	"be_uas/config"
	"be_uas/database"
	"log"
	"os"
	_ "be_uas/docs"
)

// @title           Student Achievement Reporting API
// @version         1.0
// @description     API untuk pelaporan dan verifikasi prestasi mahasiswa.
// @termsOfService  http://swagger.io/terms/

// @contact.name    Tim Pengembang
// @contact.email   support@univ.ac.id

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:3000
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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