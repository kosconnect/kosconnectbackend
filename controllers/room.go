package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
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
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Mengubah angka ke format rupiah
func formatrupiah(price float64) string {
	formatter := message.NewPrinter(language.Indonesian)
	return formatter.Sprintf("Rp %.2f", price)
}

func CreateRoom(c *gin.Context) {
	// Parse form-data
	err := c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form-data"})
		return
	}

	// Extract fields
	boardingHouseID, _ := primitive.ObjectIDFromHex(c.PostForm("boarding_house_id"))
	roomType := c.PostForm("room_type")
	size := c.PostForm("size")

	// Parsing and defaulting prices
	priceMonthlyStr := c.PostForm("price_monthly")
	priceMonthly := 0
	if priceMonthlyStr != "" {
		priceMonthly, _ = strconv.Atoi(priceMonthlyStr)
	}

	priceQuarterlyStr := c.PostForm("price_quarterly")
	priceQuarterly := 0
	if priceQuarterlyStr != "" {
		priceQuarterly, _ = strconv.Atoi(priceQuarterlyStr)
	}

	priceSemiAnnualStr := c.PostForm("price_semi_annual")
	priceSemiAnnual := 0
	if priceSemiAnnualStr != "" {
		priceSemiAnnual, _ = strconv.Atoi(priceSemiAnnualStr)
	}

	priceYearlyStr := c.PostForm("price_yearly")
	priceYearly := 0
	if priceYearlyStr != "" {
		priceYearly, _ = strconv.Atoi(priceYearlyStr)
	}

	// Parse Room Facilities
	var roomFacilities []models.RoomFacilities
	roomFacilitiesJSON := c.PostForm("room_facilities")
	if err := json.Unmarshal([]byte(roomFacilitiesJSON), &roomFacilities); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room facilities format"})
		return
	}

	// Parse Custom Facilities
	var customFacilities []models.CustomFacilities
	customFacilitiesJSON := c.PostForm("custom_facilities")
	if err := json.Unmarshal([]byte(customFacilitiesJSON), &customFacilities); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom facilities format"})
		return
	}

	// Process images
	var roomImageURL []string
	form, err := c.MultipartForm()
	if err == nil {
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
			uniqueFilename := fmt.Sprintf("RoomImages/%s%s", uuid.New().String(), ext)

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

			roomImageURL = append(roomImageURL, imageURL)
		}
	}
	// Ambil nilai dari form (atau set default jika kosong)
	numberAvailableStr := c.PostForm("number_available")
	numberAvailable := 0
	if numberAvailableStr != "" {
		numberAvailable, _ = strconv.Atoi(numberAvailableStr)
	}

	status := "Tidak Tersedia"
	if numberAvailable >= 1 {
		status = "Tersedia"
	}

	// Create room model
	room := models.Room{
		ID:              primitive.NewObjectID(),
		BoardingHouseID: boardingHouseID,
		RoomType:        roomType,
		Size:            size,
		Price: models.RoomPrice{
			Monthly:    priceMonthly,
			Quarterly:  priceQuarterly,
			SemiAnnual: priceSemiAnnual,
			Yearly:     priceYearly,
		},
		RoomFacilities:   roomFacilities,
		CustomFacilities: customFacilities,
		NumberAvailable:  numberAvailable,
		Status:           status,
		Images:           roomImageURL,
	}

	collection := config.DB.Collection("rooms")
	_, err = collection.InsertOne(context.Background(), room)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save room to database"})
		return
	}

	// Format response prices to rupiah
	formattedPrices := map[string]string{
		"monthly":     formatrupiah(float64(priceMonthly)),
		"quarterly":   formatrupiah(float64(priceQuarterly)),
		"semi_annual": formatrupiah(float64(priceSemiAnnual)),
		"yearly":      formatrupiah(float64(priceYearly)),
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Room created successfully",
		"data": gin.H{
			"room":             room,
			"prices_formatted": formattedPrices,
		},
	})
}

// GetAllRoom retrieves all rooms for public view
func GetAllRoom(c *gin.Context) {
	collection := config.DB.Collection("rooms")
	var rooms []models.Room

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rooms"})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var room models.Room
		if err := cursor.Decode(&room); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding room"})
			return
		}
		rooms = append(rooms, room)
	}

	c.JSON(http.StatusOK, gin.H{"data": rooms})
}

// ambil data rooms berdasarkan id boardinghouse
func GetRoomByBoardingHouseID(c *gin.Context) {
	// Ambil klaim JWT dari context
	claims := c.MustGet("user").(jwt.MapClaims)

	// Validasi role (opsional, jika hanya owner yang boleh akses)
	if role, ok := claims["role"].(string); !ok || role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners can view rooms"})
		return
	}

	// Ambil owner_id dari klaim
	ownerID := claims["user_id"].(string)
	ownerObjectID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid owner ID"})
		return
	}

	// Ambil boarding_house_id dari parameter URL
	boardingHouseID := c.Param("id")
	boardingHouseObjectID, err := primitive.ObjectIDFromHex(boardingHouseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid boarding house ID"})
		return
	}

	// Validasi apakah boarding house milik owner
	boardingHouseCollection := config.DB.Collection("boardinghouses")
	filter := bson.M{"_id": boardingHouseObjectID, "owner_id": ownerObjectID}
	count, err := boardingHouseCollection.CountDocuments(context.Background(), filter)
	if err != nil || count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Boarding house not found or not authorized"})
		return
	}

	// Ambil semua kamar berdasarkan boarding_house_id
	roomCollection := config.DB.Collection("rooms")
	var rooms []models.Room
	cursor, err := roomCollection.Find(context.Background(), bson.M{"boarding_house_id": boardingHouseObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rooms"})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var room models.Room
		if err := cursor.Decode(&room); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding room"})
			return
		}
		rooms = append(rooms, room)
	}

	// Kirim data rooms ke frontend
	c.JSON(http.StatusOK, gin.H{"data": rooms})
}

// GetRoomByID retrieves a specific room by ID
func GetRoomByID(c *gin.Context) {
	id := c.Param("id")
	roomID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	collection := config.DB.Collection("rooms")
	var room models.Room

	err = collection.FindOne(context.Background(), bson.M{"_id": roomID}).Decode(&room)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": room})
}

func GetRoomDetailByID(c *gin.Context) {
	roomID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	roomCollection := config.DB.Collection("rooms")

	// Pipeline untuk menggabungkan data dan gambar
	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: bson.D{
				{Key: "_id", Value: objectID}, // Filter berdasarkan Room ID
			}},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "boardinghouses"},          // Gabungkan dengan koleksi BoardingHouse
				{Key: "localField", Value: "boarding_house_id"}, // Field referensi dari koleksi Room
				{Key: "foreignField", Value: "_id"},             // Field referensi di koleksi BoardingHouse
				{Key: "as", Value: "boarding_house"},            // Hasil join disimpan dalam field boarding_house
			}},
		},
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$boarding_house"}, // Unwind untuk mengubah array menjadi objek
			}},
		},
		{
			{Key: "$addFields", Value: bson.D{
				{Key: "all_images", Value: bson.D{
					{Key: "$concatArrays", Value: bson.A{"$images", "$boarding_house.images"}}, // Gabungkan gambar
				}},
			}},
		},
	}

	// Eksekusi pipeline
	cursor, err := roomCollection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}

	// Decode hasil query
	var roomDetails []bson.M
	if err := cursor.All(context.TODO(), &roomDetails); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode data"})
		return
	}

	// Kirimkan response JSON
	c.JSON(http.StatusOK, roomDetails)
}

func GetRoomsForLandingPage(c *gin.Context) {
	roomCollection := config.DB.Collection("rooms")

	pipeline := mongo.Pipeline{
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "boardinghouses"},          // Join dengan koleksi BoardingHouse
				{Key: "localField", Value: "boarding_house_id"}, // Field referensi dari koleksi Room
				{Key: "foreignField", Value: "_id"},             // Field referensi di BoardingHouse
				{Key: "as", Value: "boarding_house"},            // Hasil join disimpan di boarding_house
			}},
		},
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$boarding_house"}, // Unwind array ke objek
			}},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "categories"},                       // Join dengan koleksi Categories
				{Key: "localField", Value: "boarding_house.category_id"}, // Referensi kategori
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "category"},
			}},
		},
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$category"}, // Unwind array ke objek
			}},
		},
		{
			{Key: "$project", Value: bson.D{
				{Key: "room_name", Value: "$room_type"},                     // Nama kamar
				{Key: "boarding_house_name", Value: "$boarding_house.name"}, // Nama kos
				{Key: "address", Value: "$boarding_house.address"},          // Alamat kos
				{Key: "price", Value: "$price.monthly"},                     // Harga bulanan
				{Key: "status", Value: bson.D{ // Hitung Status
					{Key: "status", Value: bson.D{
						{Key: "$cond", Value: bson.A{
							bson.D{{Key: "$gt", Value: bson.A{"$number_available", 0}}},
							bson.D{{Key: "$concat", Value: bson.A{
								bson.D{{Key: "$toString", Value: "$number_available"}},
								" Kamar Tersedia",
							}}},
							"Tidak Tersedia",
						}},
					}},
				}},
				{Key: "images", Value: bson.D{
					{Key: "$slice", Value: bson.A{"$images", 1}}, // Gambar pertama
				}},
			}},
		},
	}

	cursor, err := roomCollection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}

	var results []bson.M
	if err := cursor.All(context.TODO(), &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode data"})
		return
	}

	c.JSON(http.StatusOK, results)
}

func UpdateRoom(c *gin.Context) {
	roomID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	// Parse form-data
	err = c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form-data"})
		return
	}

	// Extract fields
	roomType := c.PostForm("room_type")
	size := c.PostForm("size")

	priceMonthlyStr := c.PostForm("price_monthly")
	priceMonthly := 0
	if priceMonthlyStr != "" {
		priceMonthly, _ = strconv.Atoi(priceMonthlyStr)
	}

	priceQuarterlyStr := c.PostForm("price_quarterly")
	priceQuarterly := 0
	if priceQuarterlyStr != "" {
		priceQuarterly, _ = strconv.Atoi(priceQuarterlyStr)
	}

	priceSemiAnnualStr := c.PostForm("price_semi_annual")
	priceSemiAnnual := 0
	if priceSemiAnnualStr != "" {
		priceSemiAnnual, _ = strconv.Atoi(priceSemiAnnualStr)
	}

	priceYearlyStr := c.PostForm("price_yearly")
	priceYearly := 0
	if priceYearlyStr != "" {
		priceYearly, _ = strconv.Atoi(priceYearlyStr)
	}

	numberAvailableStr := c.PostForm("number_available")
	numberAvailable := 0
	if numberAvailableStr != "" {
		numberAvailable, _ = strconv.Atoi(numberAvailableStr)
	}

	var roomFacilities []models.RoomFacilities
	roomFacilitiesJSON := c.PostForm("room_facilities")
	if err := json.Unmarshal([]byte(roomFacilitiesJSON), &roomFacilities); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room facilities format"})
		return
	}

	var customFacilities []models.CustomFacilities
	customFacilitiesJSON := c.PostForm("custom_facilities")
	if err := json.Unmarshal([]byte(customFacilitiesJSON), &customFacilities); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom facilities format"})
		return
	}

	form, err := c.MultipartForm()
	updateFields := bson.M{
		"room_type":         roomType,
		"size":              size,
		"price.monthly":     priceMonthly,
		"price.quarterly":   priceQuarterly,
		"price.semi_annual": priceSemiAnnual,
		"price.yearly":      priceYearly,
		"room_facilities":   roomFacilities,
		"custom_facilities": customFacilities,
		"status":            c.PostForm("status"),
		"number_available":  numberAvailable,
	}

	// Update images
	if err == nil {
		var roomImageURL []string
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
			uniqueFilename := fmt.Sprintf("RoomImages/%s%s", uuid.New().String(), ext)

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

			roomImageURL = append(roomImageURL, imageURL)
		}

		// Ganti gambar lama dengan gambar baru
		updateFields["images"] = roomImageURL
	}
    
    status := "Tidak Tersedia"
    if numberAvailable >= 1 {
        status = "Tersedia"
    }
    updateFields["status"] = status
    
	collection := config.DB.Collection("rooms")
	result, err := collection.UpdateOne(context.Background(), bson.M{"_id": roomID}, bson.M{"$set": updateFields})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room in database"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room updated successfully"})
}

// DeleteRoom deletes an existing room by ID
func DeleteRoom(c *gin.Context) {
	id := c.Param("id")
	roomID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	collection := config.DB.Collection("rooms")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": roomID})
	if err != nil || result.DeletedCount == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete room or room not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room deleted successfully"})
}
