package database

import (
	"database/sql"
	"fmt"
	"go_ecommerce/internal/config"
	"log"

	_ "github.com/lib/pq" // Postgres driver
)

var DB *sql.DB

// Init initializes the database connection
func Init() {
	var err error

	config := config.LoadConfig()
	// Define the PostgreSQL connection string
	connStr := fmt.Sprintf("user=%v password=%v dbname=%v host=%v sslmode=%v",
		config.DBUser, config.DBPass, config.DBName, config.DBHost, config.DBSslMode)

	// Open a database connection
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}

	// Check if the connection is alive
	err = DB.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v\n", err)
	}

	log.Println("Successfully connected to the database!")

}

// CloseDB closes the database connection
func CloseDB() {
	err := DB.Close()
	if err != nil {
		log.Printf("Error closing the database: %v\n", err)
	} else {
		log.Println("Database connection closed.")
	}
}
