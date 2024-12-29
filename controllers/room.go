package controllers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/organisasi/kosconnectbackend/config"
	"github.com/organisasi/kosconnectbackend/helper"
	"github.com/organisasi/kosconnectbackend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
    priceMonthly, _ := strconv.Atoi(c.PostForm("price_monthly"))
    priceQuarterly, _ := strconv.Atoi(c.PostForm("price_quarterly"))
    priceSemiAnnual, _ := strconv.Atoi(c.PostForm("price_semi_annual"))
    priceYearly, _ := strconv.Atoi(c.PostForm("price_yearly"))
    facilityIDs := c.PostFormArray("room_facilities")
    customFacilityIDs := c.PostFormArray("custom_facilities")
    status := c.PostForm("status")
    numberAvailable, _ := strconv.Atoi(c.PostForm("number_available"))

    // Validasi dan fetch RoomFacility IDs
    var roomFacilities []primitive.ObjectID
    roomFacilityCollection := config.DB.Collection("roomfacilities")
    for _, id := range facilityIDs {
        facilityID, err := primitive.ObjectIDFromHex(id)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid Room Facility ID: %s", id)})
            return
        }

        count, err := roomFacilityCollection.CountDocuments(context.Background(), bson.M{"_id": facilityID})
        if err != nil || count == 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Room Facility not found for ID: %s", id)})
            return
        }

        roomFacilities = append(roomFacilities, facilityID)
    }

    // Validasi dan fetch CustomFacility IDs
    var customFacilities []primitive.ObjectID
    customFacilityCollection := config.DB.Collection("customfacilities")
    for _, id := range customFacilityIDs {
        customFacilityID, err := primitive.ObjectIDFromHex(id)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid Custom Facility ID: %s", id)})
            return
        }

        count, err := customFacilityCollection.CountDocuments(context.Background(), bson.M{"_id": customFacilityID})
        if err != nil || count == 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Custom Facility not found for ID: %s", id)})
            return
        }

        customFacilities = append(customFacilities, customFacilityID)
    }

    // Process images
    var imageUrls []string
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

            imageUrls = append(imageUrls, resp.GetContent().GetHTMLURL())
        }
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
        Status:           status,
        NumberAvailable:  numberAvailable,
        Images:           imageUrls,
    }

    collection := config.DB.Collection("rooms")
    _, err = collection.InsertOne(context.Background(), room)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save room to database"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "Room created successfully",
        "data":    room,
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

// GetRoomByBoardingHouseID retrieves all rooms by boarding house ID
func GetRoomByBoardingHouseID(c *gin.Context) {
	id := c.Param("boarding_house_id")
	boardingHouseID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid boarding house ID"})
		return
	}

	collection := config.DB.Collection("rooms")
	var rooms []models.Room

	cursor, err := collection.Find(context.Background(), bson.M{"boarding_house_id": boardingHouseID})
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

// UpdateRoom updates an existing room by ID
func UpdateRoom(c *gin.Context) {
	id := c.Param("id")
	roomID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	var roomUpdates models.Room
	if err := c.ShouldBindJSON(&roomUpdates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	collection := config.DB.Collection("rooms")
	result, err := collection.UpdateOne(context.Background(), bson.M{"_id": roomID}, bson.M{"$set": roomUpdates})
	if err != nil || result.MatchedCount == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room or room not found"})
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