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
	facilityTypeIDs := c.PostFormArray("facility_type_ids")

	// Validasi dan fetch facility IDs
	var facilityRefs []primitive.ObjectID
	facilityCollection := config.DB.Collection("facilitytypes")
	for _, id := range facilityTypeIDs {
		facilityID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid facility type ID: %s", id)})
			return
		}

		count, err := facilityCollection.CountDocuments(context.Background(), bson.M{"_id": facilityID})
		if err != nil || count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Facility type not found for ID: %s", id)})
			return
		}

		facilityRefs = append(facilityRefs, facilityID)
	}

	// Dapatkan koordinat berdasarkan alamat
	latitude, longitude, err := helper.GetCoordinates(address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to get coordinates: %v", err)})
		return
	}

	// Ambil tempat publik terdekat (misalnya dalam radius 1km)
	closestPlaces, err := helper.GetClosestPlaces(longitude, latitude, config.GetHereAPIKey(), "point_of_interest", 1.0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get closest places: %v", err)})
		return
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
		ID:            primitive.NewObjectID(),
		OwnerID:       ownerObjectID,
		CategoryID:    categoryID,
		Name:          c.PostForm("name"),
		Slug:          slug,
		Address:       address,
		Longitude:     longitude,
		Latitude:      latitude,
		Description:   description,
		Facilities:    facilityRefs, // Sesuaikan input fasilitas
		Images:        imageUrls,             // Sesuaikan input gambar
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

// UPDATE
func UpdateBoardingHouse(c *gin.Context) {
    claims := c.MustGet("user").(jwt.MapClaims)

    if role, ok := claims["role"].(string); !ok || role != "owner" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only owners can update boarding houses"})
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

    var updateData models.BoardingHouse
    if err := c.ShouldBindJSON(&updateData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    collection := config.DB.Collection("boardinghouses")
    result, err := collection.UpdateOne(
        context.Background(),
        bson.M{"_id": boardingHouseID, "owner_id": ownerObjectID},
        bson.M{"$set": updateData},
    )
    if err != nil || result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Boarding house not found or you are not the owner"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Boarding house updated successfully"})
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
