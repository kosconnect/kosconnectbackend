package controllers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/organisasi/kosconnectbackend/config"
	"github.com/organisasi/kosconnectbackend/helper"
	"github.com/organisasi/kosconnectbackend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateBoardingHouse handles the creation of a new boarding house
func CreateBoardingHouse(c *gin.Context) {
	claims := c.MustGet("user").(jwt.MapClaims)

	// Validasi role harus "owner"
	if role, ok := claims["role"].(string); !ok || role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners can create boarding houses"})
		return
	}

	ownerID := claims["user_id"].(string)

	// Parse form-data
	err := c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form-data"})
		return
	}

	// Initialize MongoDB collections
	collection := config.DB.Collection("boardinghouses")
	facilityCollection := config.DB.Collection("facilitytypes")

	// Extract fields from form-data
	name := c.PostForm("name")
	address := c.PostForm("address")
	description := c.PostForm("description")
	rules := c.PostForm("rules")
	categoryID, _ := primitive.ObjectIDFromHex(c.PostForm("category_id"))
	facilityTypeIDs := c.PostFormArray("facility_type_ids")

	// Validate and fetch facility IDs
	var facilityRefs []primitive.ObjectID
	for _, id := range facilityTypeIDs {
		facilityID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid facility type ID: %s", id)})
			return
		}

		// Verify the facility exists
		count, err := facilityCollection.CountDocuments(context.Background(), bson.M{"_id": facilityID})
		if err != nil || count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Facility type not found for ID: %s", id)})
			return
		}

		facilityRefs = append(facilityRefs, facilityID)
	}

	// Process images
	var imageUrls []string
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form-data"})
		return
	}

	files := form.File["images"]
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
			return
		}
		defer file.Close()

		// Read file content
		content, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content"})
			return
		}

		// Ekstrak ekstensi file
		ext := path.Ext(fileHeader.Filename) // Misalnya ".jpg"

		// Buat nama file unik
		uniqueFilename := fmt.Sprintf("BoardingHouseImages/%s%s", uuid.New().String(), ext)

		// Configure GitHub upload
		githubConfig := helper.GitHubConfig{
			AccessToken: config.GetGitHubToken(),
			AuthorName:  "Balqis Rosa Sekamayang",
			AuthorEmail: "balqisrosasekamayang@gmail.com",
			Org:         "kosconnect",
			Repo:        "img",
			FilePath:    uniqueFilename,
			FileContent: content,
			Replace:     true,
		}

		// Upload file to GitHub
		resp, err := helper.UploadFile(context.Background(), githubConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload file to GitHub: %v", err)})
			return
		}

		// Add URL to images array
		imageUrls = append(imageUrls, resp.GetContent().GetHTMLURL())
	}

	// Generate slug
	slug := generateSlug(name)

	ownerObjectID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID"})
		return
	}

	// Create boarding house model
	boardingHouse := models.BoardingHouse{
		ID:          primitive.NewObjectID(),
		OwnerID:     ownerObjectID,
		CategoryID:  categoryID,
		Name:        name,
		Slug:        slug,
		Address:     address,
		Description: description,
		Facilities:  facilityRefs,
		Images:      imageUrls,
		Rules:       rules,
	}

	// Insert into MongoDB
	_, err = collection.InsertOne(context.Background(), boardingHouse)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save boarding house to database"})
		return
	}

	// Respond success
	c.JSON(http.StatusCreated, gin.H{
		"message": "Boarding house created successfully",
		"data":    boardingHouse,
	})
}

// generateSlug generates a URL-friendly slug from the given name
func generateSlug(name string) string {
	// Replace spaces with hyphens, remove non-alphanumeric characters, and convert to lowercase
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-]+`)
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	return reg.ReplaceAllString(slug, "") + "-" + uuid.NewString()
}
