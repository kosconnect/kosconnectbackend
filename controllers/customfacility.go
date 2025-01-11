package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/organisasi/kosconnectbackend/config"
	"github.com/organisasi/kosconnectbackend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Create CustomFacility
func CreateCustomFacility(c *gin.Context) {
	claims := c.MustGet("user").(jwt.MapClaims)
	role, _ := claims["role"].(string)

	var facility models.CustomFacility
	if err := c.ShouldBindJSON(&facility); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if role == "admin" {
		if facility.OwnerID == primitive.NilObjectID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "OwnerID is required for admin"})
			return
		}
	} else if role == "owner" {
		ownerID, _ := primitive.ObjectIDFromHex(claims["user_id"].(string))
		facility.OwnerID = ownerID
	}

	facility.CustomFacilityID = primitive.NewObjectID()
	collection := config.DB.Collection("customFacility")
	if _, err := collection.InsertOne(context.TODO(), facility); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create custom facility"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Custom facility created successfully",
		"data":    facility,
	})
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

	if err := cursor.All(context.TODO(), &facilities); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse custom facilities"})
		return
	}

	c.JSON(http.StatusOK, facilities)
}

// Get CustomFacility by ID
func GetCustomFacilityByID(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facility ID"})
		return
	}

	collection := config.DB.Collection("customFacility")
	var facility models.CustomFacility
	if err := collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&facility); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Custom facility not found"})
		return
	}

	c.JSON(http.StatusOK, facility)
}

// Get CustomFacilities by OwnerID
func GetCustomFacilitiesByOwnerID(c *gin.Context) {
	// Ambil klaim user dari JWT
	claims := c.MustGet("user").(jwt.MapClaims)

	if role, ok := claims["role"].(string); !ok || role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners can access their custom facilities"})
		return
	}

	ownerID, err := primitive.ObjectIDFromHex(claims["user_id"].(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID"})
		return
	}

	collection := config.DB.Collection("customFacility")
	var facilities []models.CustomFacility

	cursor, err := collection.Find(context.TODO(), bson.M{"owner_id": ownerID})
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

// Update CustomFacility
func UpdateCustomFacility(c *gin.Context) {
	claims := c.MustGet("user").(jwt.MapClaims)
	role, _ := claims["role"].(string)

	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facility ID"})
		return
	}

	var updateData models.CustomFacility
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	collection := config.DB.Collection("customFacility")
	filter := bson.M{"_id": objID}

	if role == "owner" {
		ownerID, _ := primitive.ObjectIDFromHex(claims["user_id"].(string))
		filter["owner_id"] = ownerID
	}

	update := bson.M{"$set": bson.M{"name": updateData.Name, "price": updateData.Price}}
	if _, err := collection.UpdateOne(context.TODO(), filter, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update custom facility"})
		return
	}

	var updatedFacility models.CustomFacility
	if err := collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&updatedFacility); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated facility"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Custom facility updated successfully",
		"data":    updatedFacility,
	})
}

// Delete CustomFacility
func DeleteCustomFacility(c *gin.Context) {
	claims := c.MustGet("user").(jwt.MapClaims)
	role, _ := claims["role"].(string)

	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid facility ID"})
		return
	}

	filter := bson.M{"_id": objID}
	if role == "owner" {
		ownerID, _ := primitive.ObjectIDFromHex(claims["user_id"].(string))
		filter["owner_id"] = ownerID
	}

	collection := config.DB.Collection("customFacility")
	if _, err := collection.DeleteOne(context.TODO(), filter); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete custom facility"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Custom facility deleted successfully"})
}