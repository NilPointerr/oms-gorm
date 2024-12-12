package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/keyurKalariya/OMS/cmd/oms-api/models"
	"github.com/keyurKalariya/OMS/cmd/oms-api/routes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

// Initialize database connection with schema setup
func initDB() (*gorm.DB, error) {
	// Connection string for PostgreSQL
	// connStr := "host=postgres-container-35 user=root password=root dbname=oms sslmode=disable"
	connStr := "host=localhost user=root password=root dbname=oms sslmode=disable"

	// Open a connection to the database using GORM v2
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Get the raw SQL DB connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if _, err := sqlDB.Exec(`CREATE SCHEMA IF NOT EXISTS gorm`); err != nil {
		return nil, err
	}

	// Handle both return values: sql.Result and error
	if _, err := sqlDB.Exec("SET search_path TO gorm"); err != nil {
		return nil, err
	}

	// Optionally, you can check the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	// Run AutoMigrate to automatically migrate the models
	err = db.AutoMigrate(&models.Item{}, &models.Order{}, &models.OrderItem{}, &models.User{}, &models.UserOrder{})
	if err != nil {
		return nil, err
	}

	log.Println("Connected to the PostgreSQL database using GORM v2")
	return db, nil
}

func main() {
	var err error
	// Initialize the global db variable
	db, err = initDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	if db == nil {
		log.Println("Database connection is nil after initialization!")
		return
	}

	log.Println("Database connection is initialized successfully")

	// Your Gin app setup
	r := gin.Default()
	routes.SetupRoutes(r, db) // Pass the GORM db instance to the routes

	log.Println("Server is running on port 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
