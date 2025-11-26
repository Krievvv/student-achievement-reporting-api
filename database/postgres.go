package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectPostgres() {
	dsn := os.Getenv("POSTGRES_DSN")
	var err error
	
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to open connection to Postgres:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping Postgres:", err)
	}

	fmt.Println("Connected to PostgreSQL")
}