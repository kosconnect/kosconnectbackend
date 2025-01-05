package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware sets up the Cross-Origin Resource Sharing (CORS) headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil Origin dari request
		origin := c.Request.Header.Get("Origin")
		allowedOrigins := []string{
			"https://kosconnect.github.io/",
			"https://localhost:8080/", // Tambahkan origin lainnya sesuai kebutuhan
		}

		// Cek jika origin request ada di daftar origin yang diizinkan
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle OPTIONS method
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		// Process request
		c.Next()
	}
}
