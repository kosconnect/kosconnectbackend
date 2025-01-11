package handler //ganti ke main kalau mau di run local

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	// "github.com/joho/godotenv" //digunakan hanya jika akan di run secara local
	"github.com/organisasi/kosconnectbackend/config"
	"github.com/organisasi/kosconnectbackend/middlewares"
	"github.com/organisasi/kosconnectbackend/routes"
)

func init() {	
	// Load environment variables digunakan hanya jika akan di run secara local
	// if err := godotenv.Load(); err != nil {
	// 	log.Println("No .env file found")
	// }

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
	routes.Facility(router)
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