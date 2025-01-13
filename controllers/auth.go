package controllers

import (
	"context"
	"crypto/rand"
	// "os/user"

	// "fmt"
	// "log"
	"net/http"

	// "os/user"

	// "regexp"
	"encoding/base64"
	"encoding/json"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/organisasi/kosconnectbackend/config"
	"github.com/organisasi/kosconnectbackend/helper"
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

// Register handles user registration SMTP
func Register(c *gin.Context) {
	var user models.User

	// Validasi input JSON
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Validasi email yang sudah terdaftar
	collection := config.DB.Collection("users")
	emailExists := collection.FindOne(context.TODO(), bson.M{"email": user.Email}).Err() == nil
	if emailExists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
		return
	}

	// Hash password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	// Set default values
	user.UserID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.VerifiedEmail = false  // Email belum diverifikasi
	user.IsRoleAssigned = false // Default untuk role

	// Generate verification token
	verifyToken := generateVerificationToken()

	// Menambahkan token verifikasi ke user
	user.VerificationToken = verifyToken

	// Simpan user ke database
	_, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Kirim email verifikasi
	verificationLink := "https://kosconnect-server.vercel.app/verify?token=" + verifyToken
	err = helper.SendVerificationEmail(user.Email, verificationLink, user.FullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Registration successful. Please check your email to verify your account.",
	})
}

// Fungsi untuk membuat token verifikasi unik
func generateVerificationToken() string {
	bytes := make([]byte, 32)                       // Membuat byte array dengan panjang 32
	rand.Read(bytes)                                // Mengisi byte array dengan nilai acak
	return base64.URLEncoding.EncodeToString(bytes) // Mengubah byte array menjadi string URL-safe
}

func VerifyEmail(c *gin.Context) {
    token := c.DefaultQuery("token", "")

    if token == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
        return
    }

    // Cari user berdasarkan token verifikasi
    collection := config.DB.Collection("users")
    var user models.User
    err := collection.FindOne(context.TODO(), bson.M{"verification_token": token}).Decode(&user)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
        return
    }

    // Update status email pengguna menjadi terverifikasi dan hapus token
    _, err = collection.UpdateOne(context.TODO(), bson.M{"_id": user.UserID}, bson.M{
        "$set": bson.M{"verified_email": true},
        "$unset": bson.M{"verification_token": ""},
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user verification"})
        return
    }

    // Redirect ke halaman login
    c.Redirect(http.StatusFound, "https://kosconnect.github.io/login?verified=true")
}

// registe yang sebelumnya:
// func Register(c *gin.Context) {
// 	var user models.User

// 	// Bind JSON input to user model
// 	if err := c.ShouldBindJSON(&user); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
// 		return
// 	}

// 	// Validate required fields
// 	if user.FullName == "" || user.Email == "" || user.Password == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
// 		return
// 	}

// 	// || user.PhoneNumber == ""

// 	// Validate phone number format (E.164)
// 	// phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
// 	// if !phoneRegex.MatchString(user.PhoneNumber) {
// 	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number format"})
// 	// 	return
// 	// }

// 	// Check if email or phone number already exists
// 	collection := config.DB.Collection("users")

// 	// Check for email existence
// 	emailExists := false
// 	// phoneNumberExists := false

// 	// Check if email exists
// 	err := collection.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&models.User{})
// 	if err == nil {
// 		emailExists = true
// 	}

// 	// // Check if phone number exists
// 	// err = collection.FindOne(context.TODO(), bson.M{"phonenumber": user.PhoneNumber}).Decode(&models.User{})
// 	// if err == nil {
// 	// 	phoneNumberExists = true
// 	// }

// 	// // Construct error message
// 	// if emailExists && phoneNumberExists {
// 	// 	c.JSON(http.StatusConflict, gin.H{"error": "Email and phone number already in use"})
// 	// 	return
// 	// }
// 	if emailExists {
// 		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
// 		return
// 	}
// 	// if phoneNumberExists {
// 	// 	c.JSON(http.StatusConflict, gin.H{"error": "Phone number already in use"})
// 	// 	return
// 	// }

// 	// Handle database errors other than no document found
// 	if err != mongo.ErrNoDocuments {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
// 		return
// 	}

// 	// Hash password
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
// 		return
// 	}
// 	user.Password = string(hashedPassword)

// 	// Set default role (if not provided)
// 	if user.Role == "" {
// 		user.Role = "user" // Default role
// 	}

// 	if len(user.Password) < 6 {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters long"})
// 		return
// 	}

// 	// Set user ID dan waktu
// 	user.UserID = primitive.NewObjectID()
// 	user.CreatedAt = time.Now()
// 	user.UpdatedAt = time.Now()

// 	// Insert user ke MongoDB
// 	_, err = collection.InsertOne(context.TODO(), user)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
// }

// Google OAuth Configuration
var googleOauthConfig = oauth2.Config{
	RedirectURL:  "https://kosconnect-server.vercel.app/auth/callback", // Sesuaikan dengan konfigurasi Anda
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

var oauthStateString = generateStateOauthCookie()

// Generate random state for security
func generateStateOauthCookie() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// HandleGoogleLogin redirects the user to the Google login page
func HandleGoogleLogin(c *gin.Context) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleGoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not found"})
		return
	}

	// Exchange the code for a token
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}

	// Fetch user info from Google
	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v1/userinfo?alt=json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email         string `json:"email"`
		Name          string `json:"name"`
		VerifiedEmail bool   `json:"verified_email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info: " + err.Error()})
		return
	}

	// Check if user exists in database
	var user models.User
	collection := config.DB.Collection("users")
	err = collection.FindOne(context.TODO(), bson.M{"email": userInfo.Email}).Decode(&user)

	if err == mongo.ErrNoDocuments {
		// Create new user
		newUser := models.User{
			UserID:        primitive.NewObjectID(),
			FullName:      userInfo.Name,
			Email:         userInfo.Email,
			Role:          "", // Role belum ditentukan
			VerifiedEmail: userInfo.VerifiedEmail,
		}
		_, err = collection.InsertOne(context.TODO(), newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// Redirect to role assignment page
		c.Redirect(http.StatusFound, "https://kosconnect.github.io/auth-assign-role?email="+userInfo.Email+"&id="+newUser.UserID.Hex())
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	// If role is not assigned
	if user.Role == "" {
		c.Redirect(http.StatusFound, "https://kosconnect.github.io/auth-assign-role?email="+user.Email+"&id="+user.UserID.Hex())
		return
	}

	// Redirect user based on role
	c.Redirect(http.StatusFound, "https://kosconnect.github.io/auth?email="+user.Email+"&role="+user.Role+"&id="+user.UserID.Hex())
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

func GoogleAuth(c *gin.Context) {
	var payload struct {
		Email string `json:"email" binding:"required,email"`
		Role  string `json:"role" binding:"required"`
	}

	// Validasi input payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Validasi role
	if payload.Role != "user" && payload.Role != "owner" && payload.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	// Cari user berdasarkan email
	collection := config.DB.Collection("users")
	var user models.User
	err := collection.FindOne(context.TODO(), bson.M{"email": payload.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Generate JWT token
	token, err := generateToken(user.UserID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set cookies (secure cookie untuk HTTPS)
	c.SetCookie("authToken", token, 3600*24*7, "/", "", true, true)

	// Response sukses
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"role":    user.Role,
		"token":   token,
	})
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
	token, err := generateToken(user.UserID, user.Role)
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
