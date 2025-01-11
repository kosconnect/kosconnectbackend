package controllers

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/organisasi/kosconnectbackend/config"
    "github.com/organisasi/kosconnectbackend/models"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "github.com/golang-jwt/jwt/v5"
)

// Create CustomFacility
func CreateCustomFacility(c *gin.Context) {
    // Ambil klaim user dari JWT
    claims := c.MustGet("user").(jwt.MapClaims)

    if role, ok := claims["role"].(string); !ok || role != "owner" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only owners can create custom facilities"})
        return
    }

    ownerID, err := primitive.ObjectIDFromHex(claims["user_id"].(string))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID"})
        return
    }

    // Bind JSON input ke struct
    var facility models.CustomFacility
    if err := c.ShouldBindJSON(&facility); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    // Generate ObjectID untuk facility
    facility.CustomFacilityID = primitive.NewObjectID().Hex()
    facility.OwnerID = ownerID

    // Simpan ke koleksi MongoDB
    collection := config.DB.Collection("customFacility")
    _, err = collection.InsertOne(context.TODO(), facility)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create custom facility"})
        return
    }

    // Format harga dalam rupiah
    formattedPrice := formatrupiah(facility.Price)

    // Respon sukses
    c.JSON(http.StatusCreated, gin.H{
        "message": "Custom facility created successfully",
        "data": gin.H{
            "id":          facility.CustomFacilityID,
            "owner_id":    facility.OwnerID,
            "name":        facility.Name,
            "price":       formattedPrice, // Harga dalam format Indonesia
        },
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
    // Ambil klaim user dari JWT
	claims := c.MustGet("user").(jwt.MapClaims)

    if role, ok := claims["role"].(string); !ok || role != "owner" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only owners can update custom facilities"})
        return
    }

    ownerID, err := primitive.ObjectIDFromHex(claims["user_id"].(string))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID"})
        return
    }

    id := c.Param("id")
    var updateData models.CustomFacility

    if err := c.ShouldBindJSON(&updateData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    collection := config.DB.Collection("customFacility")
    _, err = collection.UpdateOne(
        context.TODO(),
        bson.M{"_id": id, "owner_id": ownerID},
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
    // Ambil klaim user dari JWT
	claims := c.MustGet("user").(jwt.MapClaims)

    if role, ok := claims["role"].(string); !ok || role != "owner" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only owners can delete custom facilities"})
        return
    }

    ownerID, err := primitive.ObjectIDFromHex(claims["user_id"].(string))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID"})
        return
    }

    id := c.Param("id")
    collection := config.DB.Collection("customFacility")

    _, err = collection.DeleteOne(context.TODO(), bson.M{"_id": id, "owner_id": ownerID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete custom facility"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Custom facility deleted successfully"})
}
