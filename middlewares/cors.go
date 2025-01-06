package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware sets up the Cross-Origin Resource Sharing (CORS) headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowedOrigins := []string{
			"https://kosconnect.github.io",
			"http://localhost:8080", // Testing dengan localhost
			"https://accounts.google.com", // Google OAuth origin
		}

		allowed := false

		// Pengecualian untuk Google OAuth endpoints
		oauthEndpoints := []string{"/auth/google/login", "/auth/google/callback"}
		for _, endpoint := range oauthEndpoints {
			if c.Request.URL.Path == endpoint {
				allowed = true
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		// Periksa apakah origin berada dalam daftar origin yang diizinkan
		if !allowed {
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		if !allowed {
			// Jika origin tidak diizinkan, beri respons atau log (opsional)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Origin not allowed"})
			return
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Tangani metode OPTIONS
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		// Lanjutkan ke handler berikutnya
		c.Next()
	}
}