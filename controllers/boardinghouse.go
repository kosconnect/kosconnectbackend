package controllers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/organisasi/kosconnectbackend/config"
	"github.com/organisasi/kosconnectbackend/helper"
	"github.com/organisasi/kosconnectbackend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateBoardingHouse handles the creation of a new boarding house
func CreateBoardingHouse(c *gin.Context) {
	// Ambil owner_id dari token JWT
	ownerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: owner_id not found in token"})
		return
	}

	ownerObjID, err := primitive.ObjectIDFromHex(ownerID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid owner_id in token"})
		return
	}

	// Parse form-data
	err = c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form-data"})
		return
	}

	// Initialize GitHub configuration
	githubConfig := ghupload.GitHubConfig{
		Token:  config.GetGitHubToken(),
		Owner:  "kosconnect",
		Repo:   "img",
		Branch: "main",
		Folder: "BoardingHouseImages",
	}

	// MongoDB setup
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to MongoDB"})
		return
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("kosconnect").Collection("boardinghouses")
	facilityCollection := client.Database("kosconnect").Collection("facilityTypes")

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
		content, err := ioutil.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content"})
			return
		}

		// Generate unique filename
		uniqueFilename := fmt.Sprintf("%s/%s", githubConfig.Folder, uuid.New().String()+fileHeader.Filename)

		// Upload file to GitHub
		err = ghupload.UploadFile(context.Background(), githubConfig, uniqueFilename, content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload file to GitHub: %v", err)})
			return
		}

		// Add URL to images array
		imageUrls = append(imageUrls, fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s", githubConfig.Owner, githubConfig.Repo, githubConfig.Branch, uniqueFilename))
	}

	// Generate slug
	slug := generateSlug(name)

	// Create boarding house model
	boardingHouse := models.BoardingHouse{
		ID:          primitive.NewObjectID(),
		OwnerID:     ownerObjID,
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
