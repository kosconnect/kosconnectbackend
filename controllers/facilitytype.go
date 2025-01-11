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

// Create FacilityType
func CreateFacilityType(c *gin.Context) {
	var facilityType models.FacilityType
	if err := c.ShouldBindJSON(&facilityType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	collection := config.DB.Collection("facilitytypes")
	facilityType.FacilityTypeID = primitive.NewObjectID()

	_, err := collection.InsertOne(context.TODO(), facilityType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create facility type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Facility type created successfully"})
}

// Get All FacilityTypes
func GetAllFacilityTypes(c *gin.Context) {
    collection := config.DB.Collection("facilitytypes")
    cursor, err := collection.Find(context.TODO(), bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch facility types"})
        return
    }
    defer cursor.Close(context.TODO())

    var facilityTypes []models.FacilityType
    if err := cursor.All(context.TODO(), &facilityTypes); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse facility types"})
        return
    }

    // Mengembalikan data sesuai struktur model
    c.JSON(http.StatusOK, facilityTypes)
}


// Get FacilityType by ID
func GetFacilityTypeByID(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facility type ID"})
		return
	}

	collection := config.DB.Collection("facilitytypes")
	var facilityType models.FacilityType
	err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&facilityType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Facility type not found"})
		return
	}

	c.JSON(http.StatusOK, facilityType)
}

// Update FacilityType
func UpdateFacilityType(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facility type ID"})
		return
	}

	var facilityType models.FacilityType
	if err := c.ShouldBindJSON(&facilityType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	collection := config.DB.Collection("facilitytypes")
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": facilityType})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update facility type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Facility type updated successfully"})
}

// Delete FacilityType
func DeleteFacilityType(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facility type ID"})
		return
	}

	collection := config.DB.Collection("facilitytypes")
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete facility type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Facility type deleted successfully"})
}
