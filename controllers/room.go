package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	// Ambil boardingHouseID dari URL
	boardingHouseIDStr := c.Param("boardingHouseID")
	boardingHouseID, err := primitive.ObjectIDFromHex(boardingHouseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid BoardingHouseID"})
		return
	}

	// Ambil ownerID dari BoardingHouse
	var boardingHouse models.BoardingHouse
	collectionBoardingHouse := config.DB.Collection("boardinghouses")
	err = collectionBoardingHouse.FindOne(context.TODO(), bson.M{"_id": boardingHouseID}).Decode(&boardingHouse)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Boarding house not found"})
		return
	}
	ownerID := boardingHouse.OwnerID // Ambil ownerID dari data BoardingHouse

	// Parse form-data
	err = c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form-data"})
		return
	}

	// Ambil field dari form-data
	roomType := c.PostForm("room_type")
	size := c.PostForm("size")

	// Parsing harga dengan default 0 jika kosong
	priceMonthly, _ := strconv.Atoi(c.DefaultPostForm("price_monthly", "0"))
	priceQuarterly, _ := strconv.Atoi(c.DefaultPostForm("price_quarterly", "0"))
	priceSemiAnnual, _ := strconv.Atoi(c.DefaultPostForm("price_semi_annual", "0"))
	priceYearly, _ := strconv.Atoi(c.DefaultPostForm("price_yearly", "0"))

	// Parse Room Facilities
	var roomFacilities []primitive.ObjectID
	roomFacilitiesJSON := c.PostForm("room_facilities")
	if err := json.Unmarshal([]byte(roomFacilitiesJSON), &roomFacilities); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room facilities format"})
		return
	}

	// Validasi Room Facilities
	collectionFacilities := config.DB.Collection("facilities")
	validRoomFacilities := []primitive.ObjectID{}

	for _, facilityID := range roomFacilities {
		err := collectionFacilities.FindOne(context.TODO(), bson.M{
			"_id":  facilityID,
			"type": "room",
		}).Err()

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Room facility with ID %s is not valid or not of type 'room'", facilityID.Hex()),
			})
			return
		}

		validRoomFacilities = append(validRoomFacilities, facilityID)
	}

	// Parse Custom Facilities
	var customFacilities []primitive.ObjectID
	customFacilitiesJSON := c.PostForm("custom_facilities")
	if err := json.Unmarshal([]byte(customFacilitiesJSON), &customFacilities); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom facilities format"})
		return
	}

	collectionCustomFacilities := config.DB.Collection("customFacility")
	validCustomFacilities := []primitive.ObjectID{}

	for _, facilityID := range customFacilities {
		err := collectionCustomFacilities.FindOne(context.TODO(), bson.M{
			"_id":      facilityID,
			"owner_id": ownerID,
		}).Err()

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Custom facility with ID %s is not valid or not owned by this user", facilityID.Hex()),
			})
			return
		}

		validCustomFacilities = append(validCustomFacilities, facilityID)
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

			imageURL := strings.Replace(resp.GetContent().GetHTMLURL(), "github.com", "raw.githubusercontent.com", 1)
			imageURL = strings.Replace(imageURL, "/blob/", "/", 1)

			roomImageURL = append(roomImageURL, imageURL)
		}
	}

	// Ambil nilai dari form (atau set default jika kosong)
	numberAvailable, _ := strconv.Atoi(c.DefaultPostForm("number_available", "0"))

	status := "Tidak Tersedia"
	if numberAvailable >= 1 {
		status = "Tersedia"
	}

	// Create room model
	room := models.Room{
		RoomID:          primitive.NewObjectID(),
		BoardingHouseID: boardingHouseID,
		RoomType:        roomType,
		Size:            size,
		Price: models.RoomPrice{
			Monthly:    priceMonthly,
			Quarterly:  priceQuarterly,
			SemiAnnual: priceSemiAnnual,
			Yearly:     priceYearly,
		},
		RoomFacilities:   validRoomFacilities,
		CustomFacilities: validCustomFacilities,
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
func GetAllRooms(c *gin.Context) {
	// Mendapatkan koleksi MongoDB
	collection := config.DB.Collection("rooms")

	// Query semua kamar
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Printf("Error fetching rooms from the database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rooms from the database"})
		return
	}
	defer cursor.Close(context.TODO())

	// Array untuk menyimpan data kamar
	var rooms []models.Room

	// Decode hasil query ke dalam array rooms
	if err := cursor.All(context.TODO(), &rooms); err != nil {
		log.Printf("Error decoding rooms: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode rooms"})
		return
	}

	// Log the rooms for debugging
	log.Printf("Rooms: %+v", rooms)

	// Return semua kamar
	c.JSON(http.StatusOK, gin.H{
		"message": "Rooms fetched successfully",
		"data":    rooms,
	})
}

// GetRoomByBoardingHouseID retrieves rooms by boarding house ID
func GetRoomByBoardingHouseID(c *gin.Context) {
	// Ambil klaim JWT dari context
	claims := c.MustGet("user").(jwt.MapClaims)

	// Ambil role dari klaim JWT
	role, ok := claims["role"].(string)
	if !ok || (role != "owner" && role != "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners or admins can view rooms"})
		return
	}

	// Ambil boarding_house_id dari parameter URL
	boardingHouseID := c.Param("id")
	boardingHouseObjectID, err := primitive.ObjectIDFromHex(boardingHouseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid boarding house ID"})
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

func GetRoomDetailsByID(c *gin.Context) {
	roomID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	roomCollection := config.DB.Collection("rooms")

	// Pipeline untuk agregasi
	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: bson.D{{Key: "_id", Value: objectID}}}, // Match the room by ID
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "boardinghouses"},
				{Key: "localField", Value: "boarding_house_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "boarding_house"},
			}}, // Join with boardinghouses collection
		},
		{
			{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$boarding_house"}}}, // Unwind the boarding house array
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "boarding_house.owner_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "owner"},
			}}, // Join with users collection to get owner details
		},
		{
			{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$owner"}}}, // Unwind the owner array
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "facilities"},
				{Key: "localField", Value: "room_facilities"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "room_facilities_details"},
			}}, // Join with facilities collection
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "customFacility"},
				{Key: "localField", Value: "custom_facilities"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "custom_facility_details"},
			}}, // Join with customFacility collection
		},
		{
			{Key: "$project", Value: bson.D{
				{Key: "room_id", Value: "$_id"},
				{Key: "boarding_house_name", Value: "$boarding_house.name"},      // Include boarding house name
				{Key: "owner_name", Value: "$owner.fullname"},                    // Include owner name
				{Key: "room_facilities", Value: "$room_facilities_details.name"}, // Include room facility names
				{Key: "custom_facility_details", Value: "$custom_facility_details"},
			}},
		},
	}

	// Eksekusi pipeline
	cursor, err := roomCollection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		log.Printf("Error during aggregation: %v", err)
		return
	}

	// Decode hasil query
	var roomDetails []bson.M
	if err := cursor.All(context.TODO(), &roomDetails); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode data"})
		log.Printf("Error decoding cursor: %v", err)
		return
	}

	// Log hasil custom_facility_details untuk debug
	if len(roomDetails) > 0 {
		log.Printf("Custom Facility Details: %v", roomDetails[0]["custom_facility_details"])
	} else {
		log.Printf("No room details found for room ID: %s", roomID)
	}

	// Kirimkan response JSON
	c.JSON(http.StatusOK, roomDetails)
}

func GetRoomDetailPages(c *gin.Context) {
	roomID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	roomCollection := config.DB.Collection("rooms")

	// Pipeline untuk menggabungkan data dan gambar, termasuk owner, kategori, fasilitas, dan custom fasilitas
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
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},                         // Gabungkan dengan koleksi Users untuk mendapatkan Owner
				{Key: "localField", Value: "boarding_house.owner_id"}, // Field referensi dari koleksi BoardingHouse
				{Key: "foreignField", Value: "_id"},                   // Field referensi di koleksi Users (Owner)
				{Key: "as", Value: "owner"},                           // Hasil join disimpan dalam field owner
			}},
		},
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$owner"}, // Unwind untuk mengubah array menjadi objek
			}},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "categories"},                       // Gabungkan dengan koleksi Categories untuk mendapatkan kategori
				{Key: "localField", Value: "boarding_house.category_id"}, // Field referensi dari boarding_house
				{Key: "foreignField", Value: "_id"},                      // Field referensi di koleksi Categories
				{Key: "as", Value: "category"},                           // Hasil join disimpan dalam field category
			}},
		},
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$category"}, // Unwind untuk mengubah array menjadi objek
			}},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "facilities"}, // Gabungkan dengan koleksi Facilities
				{Key: "localField", Value: "boarding_house.facilities_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "facilities"}, // Hasil join disimpan dalam field room_facilities
			}},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "facilities"}, // Gabungkan dengan koleksi Facilities
				{Key: "localField", Value: "room_facilities"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "room_facilities"}, // Hasil join disimpan dalam field room_facilities
			}},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "customFacility"}, // Gabungkan dengan koleksi Custom Facilities
				{Key: "localField", Value: "custom_facilities"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "custom_facilities"}, // Hasil join disimpan dalam field custom_facilities
			}},
		},
		{
			{Key: "$addFields", Value: bson.D{
				{Key: "all_images", Value: bson.D{
					{Key: "$concatArrays", Value: bson.A{"$images", "$boarding_house.images"}}, // Gabungkan gambar
				}},
				{Key: "owner_id", Value: "$boarding_house.owner_id"},
				{Key: "owner_fullname", Value: "$owner.fullname"},          // Tambahkan fullname dari owner
				{Key: "category_name", Value: "$category.name"},            // Tambahkan nama kategori
				{Key: "rules", Value: "$boarding_house.rules"},             // Tambahkan nama kategori
				{Key: "description", Value: "$boarding_house.description"}, // Tambahkan nama kategori
				{Key: "address", Value: "$boarding_house.address"},         // Tambahkan nama kategori
				{Key: "room_name", Value: bson.D{
					{Key: "$concat", Value: bson.A{"$boarding_house.name", " Tipe ", "$room_type"}}, // Gabungkan nama kos dan tipe kamar
				}},
			}},
		},
		{
			{Key: "$project", Value: bson.D{
				{Key: "room_id", Value: "$_id"},
				{Key: "boarding_house_id", Value: "$boarding_house._id"},
				{Key: "owner_id", Value: 1},
				{Key: "room_name", Value: 1},
				{Key: "all_images", Value: 1},
				{Key: "owner_fullname", Value: 1},
				{Key: "category_name", Value: 1},
				{Key: "facilities", Value: "$facilities"},
				{Key: "room_facilities", Value: "$room_facilities"},
				{Key: "custom_facilities", Value: "$custom_facilities"},
				{Key: "price", Value: "$price"},
				{Key: "description", Value: 1},
				{Key: "address", Value: 1},
				{Key: "size", Value: 1},
				{Key: "rules", Value: 1},
				{Key: "number_available", Value: 1},
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

	// Define aggregation pipeline
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
				{Key: "path", Value: "$boarding_house"},          // Unwind array ke objek
				{Key: "preserveNullAndEmptyArrays", Value: true}, // Pastikan tetap ada meskipun boarding house kosong
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
				{Key: "path", Value: "$category"},                // Unwind array ke objek
				{Key: "preserveNullAndEmptyArrays", Value: true}, // Pastikan tetap ada meskipun kategori kosong
			}},
		},
		{
			{Key: "$project", Value: bson.D{
				{Key: "room_id", Value: "$_id"}, // Tambahkan room_id
				{Key: "room_name", Value: bson.D{
					{Key: "$concat", Value: bson.A{"$boarding_house.name", " Tipe ", "$room_type"}},
				}}, // Nama kamar gabungan
				{Key: "address", Value: "$boarding_house.address"}, // Alamat kos
				{Key: "price", Value: bson.D{
					{Key: "$cond", Value: bson.D{
						{Key: "if", Value: bson.D{{Key: "$gt", Value: bson.A{"$price.quarterly", nil}}}},
						{Key: "then", Value: bson.D{
							{Key: "quarterly", Value: "$price.quarterly"},
						}},
						{Key: "else", Value: bson.D{
							{Key: "$cond", Value: bson.D{
								{Key: "if", Value: bson.D{{Key: "$gt", Value: bson.A{"$price.monthly", nil}}}},
								{Key: "then", Value: bson.D{
									{Key: "monthly", Value: "$price.monthly"},
								}},
								{Key: "else", Value: bson.D{
									{Key: "$cond", Value: bson.D{
										{Key: "if", Value: bson.D{{Key: "$gt", Value: bson.A{"$price.semi_annual", nil}}}},
										{Key: "then", Value: bson.D{
											{Key: "semi_annual", Value: "$price.semi_annual"},
										}},
										{Key: "else", Value: bson.D{
											{Key: "yearly", Value: "$price.yearly"},
										}},
									}},
								}},
							}},
						}},
					}},
				}},
				{Key: "category_name", Value: "$category.name"}, // Nama kategori
				{Key: "category_id", Value: "$category._id"},    // ID kategori
				{Key: "images", Value: bson.D{
					{Key: "$slice", Value: bson.A{"$images", 1}}, // Gambar pertama
				}},
				{Key: "status", Value: bson.D{ // Hitung Status
					{Key: "$cond", Value: bson.A{
						bson.D{{Key: "$gt", Value: bson.A{"$number_available", 0}}},
						bson.D{{Key: "$concat", Value: bson.A{
							bson.D{{Key: "$toString", Value: "$number_available"}},
							" Kamar Tersedia",
						}}},
						"Tidak Tersedia",
					}},
				}},
				{Key: "owner_id", Value: "$boarding_house.owner_id"},
			}},
		},
	}

	// Menjalankan agregasi
	cursor, err := roomCollection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}

	// Menyimpan hasil agregasi
	var results []bson.M
	if err := cursor.All(context.TODO(), &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode data"})
		return
	}

	// Mengirim data hasil agregasi ke frontend
	c.JSON(http.StatusOK, results)
}

// UPDATE
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

	// Parsing harga dengan validasi minimal satu harus diisi
	priceMonthly, _ := strconv.Atoi(c.DefaultPostForm("price_monthly", "0"))
	priceQuarterly, _ := strconv.Atoi(c.DefaultPostForm("price_quarterly", "0"))
	priceSemiAnnual, _ := strconv.Atoi(c.DefaultPostForm("price_semi_annual", "0"))
	priceYearly, _ := strconv.Atoi(c.DefaultPostForm("price_yearly", "0"))

	if priceMonthly == 0 && priceQuarterly == 0 && priceSemiAnnual == 0 && priceYearly == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Minimal satu harga harus diisi"})
		return
	}

	numberAvailable, _ := strconv.Atoi(c.DefaultPostForm("number_available", "0"))

	// Validasi Room Facilities
	var roomFacilities []primitive.ObjectID
	roomFacilitiesJSON := c.PostForm("room_facilities")
	if err := json.Unmarshal([]byte(roomFacilitiesJSON), &roomFacilities); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room facilities format"})
		return
	}

	collectionFacilities := config.DB.Collection("facilities")
	validRoomFacilities := []primitive.ObjectID{}

	for _, facilityID := range roomFacilities {
		err := collectionFacilities.FindOne(context.TODO(), bson.M{
			"_id":  facilityID,
			"type": "room",
		}).Err()

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Room facility with ID %s is not valid or not of type 'room'", facilityID.Hex()),
			})
			return
		}

		validRoomFacilities = append(validRoomFacilities, facilityID)
	}

	// Validasi Custom Facilities
	var customFacilities []primitive.ObjectID
	customFacilitiesJSON := c.PostForm("custom_facilities")
	if err := json.Unmarshal([]byte(customFacilitiesJSON), &customFacilities); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom facilities format"})
		return
	}

	collectionCustomFacilities := config.DB.Collection("customFacility")
	validCustomFacilities := []primitive.ObjectID{}

	for _, facilityID := range customFacilities {
		err := collectionCustomFacilities.FindOne(context.TODO(), bson.M{
			"_id": facilityID,
		}).Err()

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Custom facility with ID %s is not valid", facilityID.Hex()),
			})
			return
		}

		validCustomFacilities = append(validCustomFacilities, facilityID)
	}

	// Handle images
	form, err := c.MultipartForm()
	var roomImageURL []string

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

			imageURL := strings.Replace(resp.GetContent().GetHTMLURL(), "github.com", "raw.githubusercontent.com", 1)
			imageURL = strings.Replace(imageURL, "/blob/", "/", 1)

			roomImageURL = append(roomImageURL, imageURL)
		}
	}

	// Update fields
	updateFields := bson.M{
		"room_type":         roomType,
		"size":              size,
		"price":             models.RoomPrice{Monthly: priceMonthly, Quarterly: priceQuarterly, SemiAnnual: priceSemiAnnual, Yearly: priceYearly},
		"room_facilities":   validRoomFacilities,
		"custom_facilities": validCustomFacilities,
		"number_available":  numberAvailable,
		"status":            map[bool]string{true: "Tersedia", false: "Tidak Tersedia"}[numberAvailable > 0],
	}

	if len(roomImageURL) > 0 {
		updateFields["images"] = roomImageURL
	}

	collection := config.DB.Collection("rooms")
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": roomID}, bson.M{"$set": updateFields})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room updated successfully"})
}

// DeleteRoom deletes an existing room by ID
func DeleteRoom(c *gin.Context) {
	// Extract and validate the room ID
	id := c.Param("id")
	roomID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	// Delete the room document from the database
	collection := config.DB.Collection("rooms")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": roomID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete room"})
		return
	}

	// Check if the room was actually deleted
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room deleted successfully"})
}
