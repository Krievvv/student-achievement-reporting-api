package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectPostgres() {
	var err error
	dsn := os.Getenv("POSTGRES_DSN")
	
	// Open connection
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to open connection to Postgres:", err)
	}

	// Ping to verify
	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping Postgres:", err)
	}

	fmt.Println("Connected to PostgreSQL")
}