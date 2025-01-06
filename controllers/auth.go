package controllers

import (
	"context"
	"net/http"
	// "regexp"
	"encoding/json"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/organisasi/kosconnectbackend/config"
	"github.com/organisasi/kosconnectbackend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Secret key for signing JWTs
var jwtSecret = []byte("your_secret_key")

func generateToken(userID primitive.ObjectID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.Hex(),
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
		"iat":     time.Now().Unix(),                     // Issued at
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// Register handles user registration
func Register(c *gin.Context) {
	var user models.User

	// Bind JSON input to user model
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Validate required fields
	if user.FullName == "" || user.Email == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	// || user.PhoneNumber == ""

	// Validate phone number format (E.164)
	// phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	// if !phoneRegex.MatchString(user.PhoneNumber) {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number format"})
	// 	return
	// }

	// Check if email or phone number already exists
	collection := config.DB.Collection("users")

	// Check for email existence
	emailExists := false
	// phoneNumberExists := false

	// Check if email exists
	err := collection.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&models.User{})
	if err == nil {
		emailExists = true
	}

	// // Check if phone number exists
	// err = collection.FindOne(context.TODO(), bson.M{"phonenumber": user.PhoneNumber}).Decode(&models.User{})
	// if err == nil {
	// 	phoneNumberExists = true
	// }

	// // Construct error message
	// if emailExists && phoneNumberExists {
	// 	c.JSON(http.StatusConflict, gin.H{"error": "Email and phone number already in use"})
	// 	return
	// }
	if emailExists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
		return
	}
	// if phoneNumberExists {
	// 	c.JSON(http.StatusConflict, gin.H{"error": "Phone number already in use"})
	// 	return
	// }

	// Handle database errors other than no document found
	if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)

	// Set default role (if not provided)
	if user.Role == "" {
		user.Role = "user" // Default role
	}

	if len(user.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters long"})
		return
	}

	// Set user ID dan waktu
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Insert user ke MongoDB
	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

var googleOauthConfig = oauth2.Config{
	RedirectURL:  "https://kosconnect-server.vercel.app/auth/callback", // Sesuaikan dengan konfigurasi Anda
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

var oauthStateString = "random-state-string" // Bisa diganti dengan nilai dinamis untuk keamanan

// HandleGoogleLogin mengarahkan pengguna ke halaman login Google
func HandleGoogleLogin(c *gin.Context) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "select_account"))
	c.Redirect(http.StatusFound, url)
}

func HandleGoogleCallback(c *gin.Context) {
	state := c.Query("state")
	if state != oauthStateString {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not found"})
		return
	}

	// Tukar "code" dengan token Google
	token, err := googleOauthConfig.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	client := googleOauthConfig.Client(c, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v1/userinfo?alt=json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	var user models.User
	collection := config.DB.Collection("users")
	err = collection.FindOne(context.TODO(), bson.M{"email": userInfo.Email}).Decode(&user)

	if err == mongo.ErrNoDocuments {
		newUser := models.User{
			ID:            primitive.NewObjectID(),
			FullName:      userInfo.Name,
			Email:         userInfo.Email,
			Role:          "", // Role kosong
			VerifiedEmail: true,
		}
		_, err = collection.InsertOne(context.TODO(), newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// Redirect ke frontend untuk pemilihan role
		c.Redirect(http.StatusFound, "https://kosconnect.github.io/auth?email="+userInfo.Email)
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if user.Role == "" {
		c.Redirect(http.StatusFound, "https://kosconnect.github.io/auth?email="+user.Email)
		return
	}

	// Generate token dan set cookies
	tokenString, err := generateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set token sebagai cookie
	c.SetCookie(
		"authToken", // Cookie name
		tokenString, // Value
		3600*24*7,   // Expiry time in seconds (7 days)
		"/",         // Path
		"",          // Domain (empty means same as the server's domain)
		true,        // Secure (true for HTTPS only)
		true,        // HttpOnly (true prevents JavaScript access)
	)

	// Set role sebagai cookie
	c.SetCookie(
		"userRole", // Cookie name
		user.Role,  // Value
		3600*24*7,  // Expiry time in seconds (7 days)
		"/",        // Path
		"",         // Domain (empty means same as the server's domain)
		true,       // Secure (true for HTTPS only)
		false,      // HttpOnly (false to allow JavaScript access)
	)

	// Redirect setelah set cookies
	redirectURL := "https://kosconnect.github.io/"
	if user.Role == "user" {
		redirectURL = "https://kosconnect.github.io/"
	} else	if user.Role == "owner" {
		redirectURL = "https://kosconnect.github.io/dashboard-owner"
	} else if user.Role == "admin" {
		redirectURL = "https://kosconnect.github.io/dashboard-admin"
	}

	// Response JSON (Opsional, bisa diproses frontend)
	c.JSON(http.StatusOK, gin.H{
		"message":     "Login successful",
		"token":       tokenString,
		"role":        user.Role,
		"redirectURL": redirectURL, // Redirect URL dikirim
	})

	// Periksa status HTTP dan struktur JSON sebelum redirect
	if c.Writer.Status() == http.StatusOK {
		// Pastikan struktur JSON sesuai (sesuaikan dengan kebutuhan Anda)
		if _, ok := c.Get("redirectURL"); ok {
			c.Redirect(http.StatusFound, redirectURL)
		} else {
			// Handle jika tidak ada field "redirectURL" dalam JSON
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing redirectURL"})
		}
	} else {
		// Handle jika status HTTP bukan 200
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

func AssignRole(c *gin.Context) {
	var payload struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Validasi role
	if payload.Role != "user" && payload.Role != "owner" && payload.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	// Update role di database
	collection := config.DB.Collection("users")
	_, err := collection.UpdateOne(context.TODO(), bson.M{"email": payload.Email}, bson.M{"$set": bson.M{"role": payload.Role}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role assigned successfully"})
}

func Login(c *gin.Context) {
	var loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Bind JSON input
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Cari user berdasarkan email
	collection := config.DB.Collection("users")
	var user models.User
	err := collection.FindOne(context.TODO(), bson.M{"email": loginData.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Cek password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := generateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set token sebagai cookie
	c.SetCookie(
		"authToken", // Cookie name
		token,       // Value
		3600*24*7,   // Expiry time in seconds (7 days)
		"/",         // Path
		"",          // Domain (empty means same as the server's domain)
		true,        // Secure (true for HTTPS only)
		true,        // HttpOnly (true prevents JavaScript access)
	)

	// Set role sebagai cookie
	c.SetCookie(
		"userRole", // Cookie name
		user.Role,  // Value
		3600*24*7,  // Expiry time in seconds (7 days)
		"/",        // Path
		"",         // Domain (empty means same as the server's domain)
		true,       // Secure (true for HTTPS only)
		false,      // HttpOnly (false to allow JavaScript access)
	)

	// Kirim respon sukses
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"role":    user.Role,
	})
}
