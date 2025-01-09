package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strconv"
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

// CREATE / POST
func CreateBoardingHouse(c *gin.Context) {
	// Ambil klaim user dari JWT
	claims := c.MustGet("user").(jwt.MapClaims)

	// Validasi role "owner"
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

	// Extract fields from form-data
	name := c.PostForm("name")
	address := c.PostForm("address")
	description := c.PostForm("description")
	rules := c.PostForm("rules")
	categoryID, _ := primitive.ObjectIDFromHex(c.PostForm("category_id"))

	if name == "" || address == "" || description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name, address, and description are required"})
		return
	}

	if categoryID.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing category_id"})
		return
	}

	// Parse facilities
	facilitiesJSON := c.PostForm("facilities")
	var facilities []models.Facilities
	if err := json.Unmarshal([]byte(facilitiesJSON), &facilities); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facilities format"})
		return
	}

	// Ambil longitude dan latitude dari form-data
	latitudeStr := c.PostForm("latitude")
	longitudeStr := c.PostForm("longitude")
	latitude, err := strconv.ParseFloat(latitudeStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid latitude: %v", err)})
		return
	}
	longitude, err := strconv.ParseFloat(longitudeStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid longitude: %v", err)})
		return
	}

	// Ambil tempat publik terdekat (misalnya dalam radius 1km)
	maxDistance := 1.0 // Jarak maksimal dalam km, misalnya 1 km
	closestPlaces, err := helper.GetClosestPlaces(longitude, latitude, config.GetHereAPIKey(), "point_of_interest", maxDistance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get closest places: %v", err)})
		return
	}

	// Log untuk memastikan closestPlaces berisi data
	fmt.Printf("Closest places: %+v\n", closestPlaces)

	// Process images
	var boardinghouseImageURL []string
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

		content, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content"})
			return
		}

		ext := path.Ext(fileHeader.Filename)
		uniqueFilename := fmt.Sprintf("BoardingHouseImages/%s%s", uuid.New().String(), ext)

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

		resp, err := helper.UploadFile(context.Background(), githubConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload file to GitHub: %v", err)})
			return
		}

		// Ganti URL GitHub dengan raw.githubusercontent.com dan perbaiki path (hilangkan 'blob')
		imageURL := strings.Replace(resp.GetContent().GetHTMLURL(), "github.com", "raw.githubusercontent.com", 1)
		imageURL = strings.Replace(imageURL, "/blob/", "/", 1)

		boardinghouseImageURL = append(boardinghouseImageURL, imageURL)
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
		ID:            primitive.NewObjectID(),
		OwnerID:       ownerObjectID,
		CategoryID:    categoryID,
		Name:          c.PostForm("name"),
		Slug:          slug,
		Address:       address,
		Longitude:     longitude,
		Latitude:      latitude,
		Description:   description,
		Facilities:    facilities,            // Sesuaikan input fasilitas
		Images:        boardinghouseImageURL, // Sesuaikan input gambar
		Rules:         rules,
		ClosestPlaces: closestPlaces,
	}

	collection := config.DB.Collection("boardinghouses")
	_, err = collection.InsertOne(context.Background(), boardingHouse)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save boarding house to database"})
		return
	}

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

// AMBIL SEMUA BOARDING HOUSE INI BUAT DI FE LANDING PAGE
func GetAllBoardingHouse(c *gin.Context) {
	collection := config.DB.Collection("boardinghouses")
	var boardingHouses []models.BoardingHouse

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch boarding houses"})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var boardingHouse models.BoardingHouse
		if err := cursor.Decode(&boardingHouse); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding boarding house"})
			return
		}
		boardingHouses = append(boardingHouses, boardingHouse)
	}

	c.JSON(http.StatusOK, gin.H{"data": boardingHouses})
}

// GET BOARDING HOUSE BY ID (ROLE BEBAS)
func GetBoardingHouseByID(c *gin.Context) {
	claims := c.MustGet("user").(jwt.MapClaims)
	if _, ok := claims["user_id"].(string); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	id := c.Param("id")
	boardingHouseID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid boarding house ID"})
		return
	}

	collection := config.DB.Collection("boardinghouses")
	var boardingHouse models.BoardingHouse

	err = collection.FindOne(context.Background(), bson.M{"_id": boardingHouseID}).Decode(&boardingHouse)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Boarding house not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": boardingHouse,
	})
}

// INI BUAT PEMILIK KOS
func GetBoardingHouseByOwnerID(c *gin.Context) {
	claims := c.MustGet("user").(jwt.MapClaims)

	if role, ok := claims["role"].(string); !ok || role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners can view their boarding houses"})
		return
	}

	ownerID := claims["user_id"].(string)
	ownerObjectID, _ := primitive.ObjectIDFromHex(ownerID)

	collection := config.DB.Collection("boardinghouses")
	var boardingHouses []models.BoardingHouse

	cursor, err := collection.Find(context.Background(), bson.M{"owner_id": ownerObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch boarding houses"})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var boardingHouse models.BoardingHouse
		if err := cursor.Decode(&boardingHouse); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding boarding house"})
			return
		}
		boardingHouses = append(boardingHouses, boardingHouse)
	}

	c.JSON(http.StatusOK, gin.H{"data": boardingHouses})
}

// UPDATE / PATCH
func UpdateBoardingHouse(c *gin.Context) {
	// Ambil ID boarding house dari parameter
	boardingHouseID := c.Param("id")

	// Convert ID boarding house
	objectID, err := primitive.ObjectIDFromHex(boardingHouseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid boarding house ID"})
		return
	}

	// Parse form-data
	err = c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form-data"})
		return
	}

	// Extract fields to update
	updateFields := bson.M{}

	if name := c.PostForm("name"); name != "" {
		updateFields["name"] = name
		updateFields["slug"] = generateSlug(name)
	}
	if address := c.PostForm("address"); address != "" {
		updateFields["address"] = address
	}
	if description := c.PostForm("description"); description != "" {
		updateFields["description"] = description
	}
	if rules := c.PostForm("rules"); rules != "" {
		updateFields["rules"] = rules
	}
	if categoryID := c.PostForm("category_id"); categoryID != "" {
		id, err := primitive.ObjectIDFromHex(categoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
			return
		}
		updateFields["category_id"] = id
	}

	// Update facilities
	if facilitiesJSON := c.PostForm("facilities"); facilitiesJSON != "" {
		var facilities []models.Facilities
		if err := json.Unmarshal([]byte(facilitiesJSON), &facilities); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facilities format"})
			return
		}
		updateFields["facilities"] = facilities
	}

	// Update location and fetch closest places
	latitudeStr := c.PostForm("latitude")
	longitudeStr := c.PostForm("longitude")
	if latitudeStr != "" && longitudeStr != "" {
		latitude, err := strconv.ParseFloat(latitudeStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid latitude: %v", err)})
			return
		}
		longitude, err := strconv.ParseFloat(longitudeStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid longitude: %v", err)})
			return
		}

		updateFields["latitude"] = latitude
		updateFields["longitude"] = longitude

		// Fetch closest places
		maxDistance := 1.0
		closestPlaces, err := helper.GetClosestPlaces(longitude, latitude, config.GetHereAPIKey(), "point_of_interest", maxDistance)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get closest places: %v", err)})
			return
		}
		updateFields["closest_places"] = closestPlaces
	}

	// Update images
	form, err := c.MultipartForm()
	if err == nil {
		var boardinghouseImageURL []string
		files := form.File["images"]
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
				return
			}
			defer file.Close()

			content, err := io.ReadAll(file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content"})
				return
			}

			ext := path.Ext(fileHeader.Filename)
			uniqueFilename := fmt.Sprintf("BoardingHouseImages/%s%s", uuid.New().String(), ext)

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

			resp, err := helper.UploadFile(context.Background(), githubConfig)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload file to GitHub: %v", err)})
				return
			}

			// Ganti URL GitHub dengan raw.githubusercontent.com
			imageURL := strings.Replace(resp.GetContent().GetHTMLURL(), "github.com", "raw.githubusercontent.com", 1)
			imageURL = strings.Replace(imageURL, "/blob", "", 1)

			boardinghouseImageURL = append(boardinghouseImageURL, imageURL)
		}

		// Tambahkan gambar baru jika ada
		if len(boardinghouseImageURL) > 0 {
			existingImages, _ := updateFields["images"].([]string)
			updatedImages := append(existingImages, boardinghouseImageURL...)
			updateFields["images"] = updatedImages
		}
	}

	// Periksa apakah ada field yang di-update
	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Update database
	collection := config.DB.Collection("boardinghouses")
	res, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID}, // Query hanya berdasarkan ID
		bson.M{"$set": updateFields},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update boarding house"})
		return
	}

	// Validasi apakah ada dokumen yang di-update
	if res.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No boarding house found or no changes made"})
		return
	}

	// Ambil data boarding house terbaru untuk response
	var updatedBoardingHouse models.BoardingHouse
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&updatedBoardingHouse)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated boarding house"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Boarding house updated successfully",
		"data":    updatedBoardingHouse,
	})
}

// DELETE OLEH OWNER
func DeleteBoardingHouse(c *gin.Context) {
	claims := c.MustGet("user").(jwt.MapClaims)

	if role, ok := claims["role"].(string); !ok || role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners can delete boarding houses"})
		return
	}

	ownerID := claims["user_id"].(string)
	ownerObjectID, _ := primitive.ObjectIDFromHex(ownerID)

	id := c.Param("id")
	boardingHouseID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid boarding house ID"})
		return
	}

	collection := config.DB.Collection("boardinghouses")
	result, err := collection.DeleteOne(
		context.Background(),
		bson.M{"_id": boardingHouseID, "owner_id": ownerObjectID},
	)
	if err != nil || result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Boarding house not found or you are not the owner"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Boarding house deleted successfully"})
}
