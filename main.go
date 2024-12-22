package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/organisasi/kosconnectbackend/config"
	"github.com/organisasi/kosconnectbackend/routes"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Connect to MongoDB
	config.ConnectDB()

	// Set up Gin router
	router := gin.Default()

	// Register routes
	routes.AuthRoutes(router)

	// Memanggil route untuk user (create, update, delete, get)
	routes.UserRoutes(router) // Pastikan ini ada

	// Memanggil route untuk user (create, update, delete, get)
	routes.CustomFacilityRoutes(router) // Pastikan ini ada

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(router.Run(":" + port))
}
