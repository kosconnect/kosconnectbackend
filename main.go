package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/organisasi/kosconnectbackend/config"
	"github.com/organisasi/kosconnectbackend/middlewares"
	"github.com/organisasi/kosconnectbackend/routes"
)

func init() {	
	// Connect to MongoDB
	config.ConnectDB()
}

// Handler for deployment - Menerima request dan menangani routing dengan CORS
func Handler(w http.ResponseWriter, r *http.Request) {
	// Set up Gin router di dalam handler
	router := gin.Default()

	// Apply CORS Middleware
	router.Use(middlewares.CORSMiddleware())

	// Register routes setelah router diinisialisasi
	routes.AuthRoutes(router)
	routes.UserRoutes(router)
	routes.CustomFacility(router)
	routes.CategoryRoutes(router)
	routes.BoardingHouse(router)
	routes.FacilityType(router)
	routes.RoomFacility(router)
	routes.RoomRoutes(router)

	// Handle HTTP request
	router.ServeHTTP(w, r)
}

func main() {
	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server is running on port %s\n", port)

	// Gunakan handler untuk menangani request HTTP
	http.HandleFunc("/", Handler) // Memetakan path "/" ke Handler

	// Mulai server
	log.Fatal(http.ListenAndServe(":"+port, nil)) // Tidak perlu http.HandlerFunc(Handler) di sini
}