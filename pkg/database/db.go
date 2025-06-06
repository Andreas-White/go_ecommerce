package database

import (
	"database/sql"
	"fmt"
	"go_ecommerce/internal/config"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init(cfg *config.Config) (*sql.DB, error) {
	var err error

	connStr := fmt.Sprintf("user=%v password=%v dbname=%v host=%v sslmode=%v",
		cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBHost, cfg.DBSslMode)

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("{database/Init - error opening database: %w}", err)
	}

	err = DB.Ping()
	if err != nil {
		return nil, fmt.Errorf("{database/Init - error connecting to the database: %w}", err)
	}

	log.Println("Successfully connected to the database!")

	return DB, nil
}

func CloseDB() {
	if DB != nil {
		err := DB.Close()
		if err != nil {
			log.Printf("Error closing the database: %v\n", err)
		}
		log.Println("Database connection closed.")
	}
}

func RunMigrations(cfg *config.Config) error {
	migrationDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSslMode)

	migrationsPath := "file://migrations"

	log.Printf("Attempting to run migrations from: %s on database: %s", migrationsPath, cfg.DBName)

	m, err := migrate.New(migrationsPath, migrationDSN)
	if err != nil {
		return fmt.Errorf("database.RunMigrations - failed to create new migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No new migrations to apply.")
			return nil
		}
		return fmt.Errorf("database.RunMigrations - failed to apply migrations: %w", err)
	}

	log.Println("Migrations applied successfully.")
	return nil
}
