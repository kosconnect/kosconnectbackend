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

// Create CustomFacility
func CreateCustomFacility(c *gin.Context) {
    var facility models.CustomFacility

    if err := c.ShouldBindJSON(&facility); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    facility.ID = primitive.NewObjectID().Hex()
    collection := config.DB.Collection("customFacility")

    _, err := collection.InsertOne(context.TODO(), facility)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create custom facility"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "Custom facility created successfully", "data": facility})
}

// Get All CustomFacilities
func GetAllCustomFacilities(c *gin.Context) {
    collection := config.DB.Collection("customFacility")
    var facilities []models.CustomFacility

    cursor, err := collection.Find(context.TODO(), bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch custom facilities"})
        return
    }
    defer cursor.Close(context.TODO())

    for cursor.Next(context.TODO()) {
        var facility models.CustomFacility
        if err := cursor.Decode(&facility); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding custom facility"})
            return
        }
        facilities = append(facilities, facility)
    }

    c.JSON(http.StatusOK, facilities)
}

// Get CustomFacility by ID
func GetCustomFacilityByID(c *gin.Context) {
    id := c.Param("id")
    collection := config.DB.Collection("customFacility")

    var facility models.CustomFacility
    err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&facility)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Custom facility not found"})
        return
    }

    c.JSON(http.StatusOK, facility)
}

// Update CustomFacility
func UpdateCustomFacility(c *gin.Context) {
    id := c.Param("id")
    var updateData models.CustomFacility

    if err := c.ShouldBindJSON(&updateData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    collection := config.DB.Collection("customFacility")
    _, err := collection.UpdateOne(
        context.TODO(),
        bson.M{"_id": id},
        bson.M{"$set": bson.M{"name": updateData.Name, "price": updateData.Price}},
    )

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update custom facility"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Custom facility updated successfully"})
}

// Delete CustomFacility
func DeleteCustomFacility(c *gin.Context) {
    id := c.Param("id")
    collection := config.DB.Collection("customFacility")

    _, err := collection.DeleteOne(context.TODO(), bson.M{"_id": id})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete custom facility"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Custom facility deleted successfully"})
}
