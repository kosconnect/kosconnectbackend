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
	"go.mongodb.org/mongo-driver/mongo"
)

// CREATE / POST
func CreateBoardingHouse(c *gin.Context) {
	// Ambil klaim user dari JWT
	claims := c.MustGet("user").(jwt.MapClaims)

	// Ambil role dan user ID dari klaim
	role, ok := claims["role"].(string)
	if !ok || (role != "admin" && role != "owner") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins or owners can create boarding houses"})
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

	// Validasi OwnerID untuk admin
	var ownerObjectID primitive.ObjectID
	if role == "admin" {
		ownerIDFromForm := c.PostForm("owner_id")
		if ownerIDFromForm == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "OwnerID is required for admin"})
			return
		}

		var err error
		ownerObjectID, err = primitive.ObjectIDFromHex(ownerIDFromForm)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OwnerID format"})
			return
		}
	} else if role == "owner" {
		var err error
		ownerObjectID, err = primitive.ObjectIDFromHex(ownerID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OwnerID format"})
			return
		}
	}

	// Validasi fasilitas
	facilitiesJSON := c.PostForm("facilities")
	var facilitiesIDs []string
	if err := json.Unmarshal([]byte(facilitiesJSON), &facilitiesIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facilities format"})
		return
	}

	// Validasi setiap fasilitas di database
	collectionFacilities := config.DB.Collection("facilities")
	validFacilities := []primitive.ObjectID{}
	for _, facilityID := range facilitiesIDs {
		facilityObjectID, err := primitive.ObjectIDFromHex(facilityID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facility ID"})
			return
		}

		var facility models.CustomFacility
		err = collectionFacilities.FindOne(context.TODO(), bson.M{
			"_id":  facilityObjectID,
			"type": "boarding_house",
		}).Decode(&facility)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facility type or facility does not exist"})
			return
		}

		validFacilities = append(validFacilities, facilityObjectID)
	}

	// Ambil longitude dan latitude
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

	// Proses gambar
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

		imageURL := strings.Replace(resp.GetContent().GetHTMLURL(), "github.com", "raw.githubusercontent.com", 1)
		imageURL = strings.Replace(imageURL, "/blob/", "/", 1)

		boardinghouseImageURL = append(boardinghouseImageURL, imageURL)
	}

	// Generate slug
	slug := generateSlug(name)

	// Buat model boarding house
	boardingHouse := models.BoardingHouse{
		BoardingHouseID: primitive.NewObjectID(),
		OwnerID:         ownerObjectID,
		CategoryID:      categoryID,
		Name:            name,
		Slug:            slug,
		Address:         address,
		Longitude:       longitude,
		Latitude:        latitude,
		Description:     description,
		Facilities:      validFacilities,
		Images:          boardinghouseImageURL,
		Rules:           rules,
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

func GetAllBoardingHouse(c *gin.Context) {
	collection := config.DB.Collection("boardinghouses")

	// Ambil semua data boarding house dari database
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch boarding houses"})
		return
	}
	defer cursor.Close(context.TODO())

	// Simpan hasilnya dalam slice
	var boardingHouses []models.BoardingHouse
	if err = cursor.All(context.TODO(), &boardingHouses); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse boarding houses"})
		return
	}

	// Kembalikan data tanpa memodifikasi
	c.JSON(http.StatusOK, gin.H{
		"message": "Boarding houses fetched successfully",
		"data":    boardingHouses,
	})
}

func GetBoardingHouseDetails(c *gin.Context) {
	boardingHouseID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(boardingHouseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid boarding house ID"})
		return
	}

	boardingHouseCollection := config.DB.Collection("boardinghouses")

	// Pipeline untuk mengambil category_id, owner_id, dan facilities
	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: bson.D{
				{Key: "_id", Value: objectID}, // Filter berdasarkan BoardingHouse ID
			}},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "categories"},            // Gabungkan dengan koleksi Categories
				{Key: "localField", Value: "category_id"},    // Field referensi dari BoardingHouse
				{Key: "foreignField", Value: "_id"},          // Field referensi di koleksi Categories
				{Key: "as", Value: "category"},               // Hasil join disimpan dalam field category
			}},
		},
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$category"}, // Unwind untuk mengubah array menjadi objek
			}},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},                 // Gabungkan dengan koleksi Users
				{Key: "localField", Value: "owner_id"},       // Field referensi dari BoardingHouse
				{Key: "foreignField", Value: "_id"},         // Field referensi di koleksi Users
				{Key: "as", Value: "owner"},                 // Hasil join disimpan dalam field owner
			}},
		},
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$owner"}, // Unwind untuk mengubah array menjadi objek
			}},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "facilities"},           // Gabungkan dengan koleksi Facilities
				{Key: "localField", Value: "facilities_id"},  // Field referensi array dari BoardingHouse
				{Key: "foreignField", Value: "_id"},         // Field referensi di koleksi Facilities
				{Key: "as", Value: "facilities"},            // Hasil join disimpan dalam field facilities
			}},
		},
		{
			{Key: "$project", Value: bson.D{
				{Key: "category_name", Value: "$category.name"}, // Ambil nama kategori
				{Key: "owner_fullname", Value: "$owner.fullname"}, // Ambil fullname owner
				{Key: "facilities", Value: "$facilities.name"},    // Ambil nama fasilitas
			}},
		},
	}

	// Eksekusi pipeline
	cursor, err := boardingHouseCollection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}

	// Decode hasil query
	var boardingHouseDetails []bson.M
	if err := cursor.All(context.TODO(), &boardingHouseDetails); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode data"})
		return
	}

	// Kirimkan response JSON
	c.JSON(http.StatusOK, boardingHouseDetails)
}

// GetBoardingHouseByID retrieves a boarding house by ID along with its associated facility names, category name, and owner name
func GetBoardingHouseByID(c *gin.Context) {
	id := c.Param("id")
	boardingHouseID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid boarding house ID"})
		return
	}

	// Retrieve the boarding house
	collection := config.DB.Collection("boardinghouses")
	var boardingHouse models.BoardingHouse
	err = collection.FindOne(context.TODO(), bson.M{"_id": boardingHouseID}).Decode(&boardingHouse)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Boarding house not found"})
		return
	}

	// Fetch the associated category name
	collectionCategories := config.DB.Collection("categories")
	var category models.Category
	err = collectionCategories.FindOne(context.TODO(), bson.M{"_id": boardingHouse.CategoryID}).Decode(&category)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Fetch the associated owner name
	collectionUsers := config.DB.Collection("users")
	var owner models.User
	err = collectionUsers.FindOne(context.TODO(), bson.M{
		"_id":  boardingHouse.OwnerID,
		"role": "owner", // Ensure the user is an owner
	}).Decode(&owner)
	if err != nil {
		owner.FullName = "Unknown" // If the owner is not found or does not have the "owner" role, set to "Unknown"
	}

	// Fetch the associated facility names
	collectionFacilities := config.DB.Collection("facilities")
	var facilityNames []string
	for _, facilityID := range boardingHouse.Facilities {
		var facility models.Facility
		err = collectionFacilities.FindOne(context.TODO(), bson.M{"_id": facilityID}).Decode(&facility)
		if err != nil {
			continue // Skip if the facility is not found
		}
		facilityNames = append(facilityNames, facility.Name)
	}

	// Respond with boarding house data including category, owner name, and facility names
	c.JSON(http.StatusOK, gin.H{
		"boardingHouse": boardingHouse,
		"category":      category.Name,  // Include the category name
		"owner":         owner.FullName, // Include the owner name
		"facilities":    facilityNames,
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
	claims := c.MustGet("user").(jwt.MapClaims)

	// Ambil role dan user_id dari klaim JWT
	role, ok := claims["role"].(string)
	if !ok || (role != "owner" && role != "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners or admins can update boarding houses"})
		return
	}

	var ownerID primitive.ObjectID
	if role == "owner" {
		// Ambil owner_id dari klaim JWT
		var err error
		ownerID, err = primitive.ObjectIDFromHex(claims["user_id"].(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID"})
			return
		}
	}

	// Ambil ID boarding house dari parameter
	boardingHouseID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(boardingHouseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid boarding house ID"})
		return
	}

	// Filter query untuk memastikan hanya owner yang memiliki boarding house tersebut dapat mengupdate
	filter := bson.M{"_id": objectID}
	if role == "owner" {
		filter["owner_id"] = ownerID
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
		var facilities []primitive.ObjectID
		if err := json.Unmarshal([]byte(facilitiesJSON), &facilities); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facilities format"})
			return
		}
		updateFields["facilities_id"] = facilities
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
		filter, // Filter berdasarkan role
		bson.M{"$set": updateFields},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update boarding house"})
		return
	}

	// Validasi apakah ada dokumen yang di-update
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No boarding house found or unauthorized"})
		return
	}

	// Ambil data boarding house terbaru untuk response
	var updatedBoardingHouse models.BoardingHouse
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&updatedBoardingHouse)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated boarding house"})
		return
	}

	// Fetch associated category, owner, and facilities
	collectionCategories := config.DB.Collection("categories")
	var category models.Category
	err = collectionCategories.FindOne(context.TODO(), bson.M{"_id": updatedBoardingHouse.CategoryID}).Decode(&category)
	if err != nil {
		category.Name = "Unknown"
	}

	collectionUsers := config.DB.Collection("users")
	var owner models.User
	err = collectionUsers.FindOne(context.TODO(), bson.M{
		"_id":  updatedBoardingHouse.OwnerID,
		"role": "owner",
	}).Decode(&owner)
	if err != nil {
		owner.FullName = "Unknown"
	}

	collectionFacilities := config.DB.Collection("facilities")
	var facilityNames []string
	for _, facilityID := range updatedBoardingHouse.Facilities {
		var facility models.Facility
		err = collectionFacilities.FindOne(context.TODO(), bson.M{"_id": facilityID}).Decode(&facility)
		if err != nil {
			continue
		}
		facilityNames = append(facilityNames, facility.Name)
	}

	// Return the updated boarding house data
	c.JSON(http.StatusOK, gin.H{
		"message": "Boarding house updated successfully",
		"data": gin.H{
			"boarding_house": updatedBoardingHouse,
			"category":       category.Name,
			"owner":          owner.FullName,
			"facilities":     facilityNames,
		},
	})
}

// DELETE OLEH OWNER DAN ADMIN
func DeleteBoardingHouse(c *gin.Context) {
	claims := c.MustGet("user").(jwt.MapClaims)

	// Validate if the user role is 'owner' or 'admin'
	role, ok := claims["role"].(string)
	if !ok || (role != "owner" && role != "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners or admins can delete boarding houses"})
		return
	}

	// Get the owner ID from the claims and validate it's a valid ObjectID
	ownerID, ok := claims["user_id"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID"})
		return
	}
	ownerObjectID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID format"})
		return
	}

	// Get the boarding house ID from the URL params and validate it's a valid ObjectID
	id := c.Param("id")
	boardingHouseID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid boarding house ID"})
		return
	}

	// If the user is an admin, they can delete any boarding house
	// If the user is an owner, they can only delete their own boarding house
	filter := bson.M{"_id": boardingHouseID}
	if role == "owner" {
		filter["owner_id"] = ownerObjectID
	}

	// Perform the deletion
	collection := config.DB.Collection("boardinghouses")
	result, err := collection.DeleteOne(
		context.Background(),
		filter,
	)
	if err != nil || result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Boarding house not found or unauthorized"})
		return
	}

	// Return success message if deletion is successful
	c.JSON(http.StatusOK, gin.H{"message": "Boarding house deleted successfully"})
}
