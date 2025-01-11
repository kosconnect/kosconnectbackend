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

// Create Facility
func CreateFacility(c *gin.Context) {
	var facility models.Facility
	if err := c.ShouldBindJSON(&facility); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Validasi field Type
	if facility.Type != "room" && facility.Type != "boarding_house" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type must be either 'room' or 'boarding_house'"})
		return
	}

	// Set ID baru untuk fasilitas
	facility.FacilityID = primitive.NewObjectID()

	// Simpan ke database
	collection := config.DB.Collection("facilities")
	_, err := collection.InsertOne(context.TODO(), facility)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create facility"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Facility created successfully", "data": facility})
}

// Get All Facilities
func GetAllFacilities(c *gin.Context) {
	collection := config.DB.Collection("facilities")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch facilities"})
		return
	}
	defer cursor.Close(context.TODO())

	var facilities []models.Facility
	if err := cursor.All(context.TODO(), &facilities); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse facilities"})
		return
	}

	c.JSON(http.StatusOK, facilities)
}

// Get Facility by ID
func GetFacilityByID(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facility ID"})
		return
	}

	collection := config.DB.Collection("facilities")
	var facility models.Facility
	err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&facility)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Facility not found"})
		return
	}

	c.JSON(http.StatusOK, facility)
}

// GetFacilitiesByType retrieves facilities by type (room or boarding_house)
func GetFacilitiesByType(c *gin.Context) {
	// Get the 'type' query parameter from the request
	facilityType := c.DefaultQuery("type", "")

	// Validate the 'type' query parameter
	if facilityType != "room" && facilityType != "boarding_house" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type must be either 'room' or 'boarding_house'"})
		return
	}

	// Query the database to get the facilities by type
	collection := config.DB.Collection("facilities")
	filter := bson.M{"type": facilityType}

	var facilities []models.Facility
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve facilities"})
		return
	}
	defer cursor.Close(context.TODO())

	// Loop through the cursor and append each document to the facilities slice
	for cursor.Next(context.TODO()) {
		var facility models.Facility
		if err := cursor.Decode(&facility); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode facility data"})
			return
		}
		facilities = append(facilities, facility)
	}

	// Check for any cursor error
	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor error"})
		return
	}

	// Return the facilities data
	c.JSON(http.StatusOK, gin.H{"data": facilities})
}

// Update Facility
func UpdateFacility(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facility ID"})
		return
	}

	var facility models.Facility
	if err := c.ShouldBindJSON(&facility); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Validasi field Type
	if facility.Type != "room" && facility.Type != "boarding_house" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type must be either 'room' or 'boarding_house'"})
		return
	}

	// Update data di database
	collection := config.DB.Collection("facilities")
	update := bson.M{"$set": facility}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": objID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update facility"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Facility updated successfully"})
}

// Delete Facility
func DeleteFacility(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facility ID"})
		return
	}

	collection := config.DB.Collection("facilities")
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete facility"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Facility deleted successfully"})
}
