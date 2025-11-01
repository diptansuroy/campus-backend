package db

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Global database variable
var DB *gorm.DB

// Connect function - connects to database
func Connect() {
	// Load environment variables from .env file
	godotenv.Load(".env")

	// Check which database to use
	dbType := os.Getenv("DB_TYPE")

	if dbType == "sqlite" || dbType == "" {
		// Use SQLite for development (easier setup)
		database, err := gorm.Open(sqlite.Open("campus.db"), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to SQLite database:", err)
		}
		DB = database
		log.Println("✅ Connected to SQLite database")
	} else {
		// Use PostgreSQL for production
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)
		database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to PostgreSQL database:", err)
		}
		DB = database
		log.Println("✅ Connected to PostgreSQL database")
	}
}
