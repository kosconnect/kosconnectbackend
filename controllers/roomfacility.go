package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/organisasi/kosconnectbackend/config"
	"github.com/organisasi/kosconnectbackend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Create RoomFacility
func CreateRoomFacility(c *gin.Context) {
	var roomFacility models.RoomFacility
	if err := c.ShouldBindJSON(&roomFacility); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	collection := config.DB.Collection("roomfacilities")
	roomFacility.RoomFacilityID = primitive.NewObjectID()

	_, err := collection.InsertOne(context.TODO(), roomFacility)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room facility"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room facility created successfully"})
}

// Get All RoomFacilities
func GetAllRoomFacilities(c *gin.Context) {
	collection := config.DB.Collection("roomfacilities")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch room facilities"})
		return
	}
	defer cursor.Close(context.TODO())

	var roomFacilities []models.RoomFacility
	if err := cursor.All(context.TODO(), &roomFacilities); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse room facilities"})
		return
	}

	c.JSON(http.StatusOK, roomFacilities)
}

// Get RoomFacility by ID
func GetRoomFacilityByID(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room facility ID"})
		return
	}

	collection := config.DB.Collection("roomfacilities")
	var roomFacility models.RoomFacility
	err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&roomFacility)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room facility not found"})
		return
	}

	c.JSON(http.StatusOK, roomFacility)
}

// Update RoomFacility
func UpdateRoomFacility(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room facility ID"})
		return
	}

	var roomFacility models.RoomFacility
	if err := c.ShouldBindJSON(&roomFacility); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	collection := config.DB.Collection("roomfacilities")
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": roomFacility})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room facility"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room facility updated successfully"})
}

// Delete RoomFacility
func DeleteRoomFacility(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room facility ID"})
		return
	}

	collection := config.DB.Collection("roomfacilities")
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete room facility"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room facility deleted successfully"})
}
