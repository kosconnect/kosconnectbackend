package controllers

import (
	"context"
	"net/http"
	"fmt"


	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/organisasi/kosconnectbackend/config"
	"github.com/organisasi/kosconnectbackend/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// Helper function to extract the user ID from JWT claims
func getUserIDFromToken(c *gin.Context) (primitive.ObjectID, error) {
	claims, _ := c.Get("user")
	if claims == nil {
		return primitive.NilObjectID, fmt.Errorf("user not authenticated")
	}
	userIDHex := claims.(jwt.MapClaims)["user_id"].(string)
	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return userID, nil
}

// Create user (admin only)
func CreateUser(c *gin.Context) {
	// Only allow admin to create a user
	claims, _ := c.Get("user")
	role := claims.(jwt.MapClaims)["role"].(string)
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can create users"})
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)
	user.UserID = primitive.NewObjectID()

	// Insert to MongoDB
	collection := config.DB.Collection("users")
	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

// Get all users (admin only)
func GetAllUsers(c *gin.Context) {
	// Only allow admin to view all users
	claims, _ := c.Get("user")
	role := claims.(jwt.MapClaims)["role"].(string)
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can view all users"})
		return
	}

	// Fetch all users from MongoDB
	collection := config.DB.Collection("users")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer cursor.Close(context.TODO())

	var users []models.User
	if err := cursor.All(context.TODO(), &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// Get user account (for the currently logged-in user)
func GetMyAccount(c *gin.Context) {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Fetch user from MongoDB using the user ID from token
	collection := config.DB.Collection("users")
	var user models.User
	err = collection.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// GetUserByID retrieves a user's details by their ID
func GetUserByID(c *gin.Context) {
    // Get user ID from URL parameter
    userIDParam := c.Param("id")
    userID, err := primitive.ObjectIDFromHex(userIDParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    // Fetch the user from MongoDB
    collection := config.DB.Collection("users")
    var user models.User
    err = collection.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&user)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"user": user})
}

//Get All Owners (Admin can use this to choose an owner)
func GetAllOwners(c *gin.Context) {
    collection := config.DB.Collection("users")
    var owners []models.User

    cursor, err := collection.Find(context.TODO(), bson.M{"role": "owner"})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch owners"})
        return
    }
    defer cursor.Close(context.TODO())

    for cursor.Next(context.TODO()) {
        var owner models.User
        if err := cursor.Decode(&owner); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding owner"})
            return
        }
        owners = append(owners, owner)
    }

    // Return only the email and _id (ownerID) to frontend
    c.JSON(http.StatusOK, owners)
}

// Get Owner by ID (Admin can use this to view owner details)
func GetOwnerByID(c *gin.Context) {
    ownerID := c.Param("id")
    objectID, err := primitive.ObjectIDFromHex(ownerID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID"})
        return
    }

    collection := config.DB.Collection("users")
    var owner models.User

    // Fetch the owner data by ID and filter to include only name and _id
    err = collection.FindOne(context.TODO(), bson.M{"_id": objectID, "role": "owner"}).Decode(&owner)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "Owner not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch owner"})
        }
        return
    }

    // Return only the _id and fullname of the owner
    c.JSON(http.StatusOK, gin.H{
        "owner_id":   owner.UserID,
        "owner_name": owner.FullName,
    })
}

// Update user (for the logged-in user or admin)
func UpdateMe(c *gin.Context) {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if the logged-in user is updating their own account or an admin is updating any user
	if userID.Hex() != c.Param("id") {
		claims, _ := c.Get("user")
		role := claims.(jwt.MapClaims)["role"].(string)
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own account"})
			return
		}
	}

	var updatedUser models.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Update user in MongoDB
	collection := config.DB.Collection("users")
	_, err = collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": userID},
		bson.M{"$set": updatedUser},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func UpdateUser(c *gin.Context) {
	// Mendapatkan user ID dari token
	loggedInUserID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Mendapatkan role dari token JWT
	claims, _ := c.Get("user")
	role := claims.(jwt.MapClaims)["role"].(string)

	// Mendapatkan user ID yang ingin diubah dari parameter URL
	targetUserID := c.Param("id")
	targetUserObjectID, err := primitive.ObjectIDFromHex(targetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Logika kontrol akses
	if role != "admin" && loggedInUserID.Hex() != targetUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this user"})
		return
	}

	// Parsing data input yang akan diupdate
	var updatedUserData models.User
	if err := c.ShouldBindJSON(&updatedUserData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Mencegah admin tidak sengaja mengubah field sensitif tertentu (misalnya password atau role)
	updateFields := bson.M{
		"fullname":    updatedUserData.FullName,
		"email":       updatedUserData.Email,
		"phonenumber": updatedUserData.PhoneNumber,
		"picture":     updatedUserData.Picture,
	}

	// Update user di MongoDB
	collection := config.DB.Collection("users")
	_, err = collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": targetUserObjectID},
		bson.M{"$set": updateFields},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// ResetPassword allows admin to reset a user's password
func ResetPassword(c *gin.Context) {
    // Only allow admin to reset passwords
    claims, _ := c.Get("user")
    role := claims.(jwt.MapClaims)["role"].(string)
    if role != "admin" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can reset passwords"})
        return
    }

    // Get user ID from request
    userIDParam := c.Param("id")
    userID, err := primitive.ObjectIDFromHex(userIDParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    var body struct {
        NewPassword string `json:"new_password"`
    }
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    // Hash the new password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    // Update password in MongoDB
    collection := config.DB.Collection("users")
    _, err = collection.UpdateOne(
        context.TODO(),
        bson.M{"_id": userID},
        bson.M{"$set": bson.M{"password": string(hashedPassword)}},
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// ChangePassword allows a logged-in user to change their password
func ChangePassword(c *gin.Context) {
    userID, err := getUserIDFromToken(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    var body struct {
        OldPassword string `json:"old_password"`
        NewPassword string `json:"new_password"`
    }
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    // Fetch user from MongoDB
    collection := config.DB.Collection("users")
    var user models.User
    err = collection.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&user)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    // Verify old password
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.OldPassword)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Old password is incorrect"})
        return
    }

    // Hash the new password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    // Update password in MongoDB
    _, err = collection.UpdateOne(
        context.TODO(),
        bson.M{"_id": userID},
        bson.M{"$set": bson.M{"password": string(hashedPassword)}},
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// UpdateUserRole allows admin to update the role of a user
func UpdateUserRole(c *gin.Context) {
    // Only allow admin to update roles
    claims, _ := c.Get("user")
    role := claims.(jwt.MapClaims)["role"].(string)
    if role != "admin" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can update roles"})
        return
    }

    // Get user ID and new role from request
    userIDParam := c.Param("id")
    userID, err := primitive.ObjectIDFromHex(userIDParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    var body struct {
        Role string `json:"role"`
    }
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    // Validate role
    validRoles := map[string]bool{"user": true, "owner": true, "admin": true}
    if !validRoles[body.Role] {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
        return
    }

    // Update role in MongoDB
    collection := config.DB.Collection("users")
    _, err = collection.UpdateOne(
        context.TODO(),
        bson.M{"_id": userID},
        bson.M{"$set": bson.M{"role": body.Role}},
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully"})
}

// DeleteUser deletes a user (self or by admin)
func DeleteUser(c *gin.Context) {
	// Get the logged-in user's ID from the token
	loggedInUserID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get the target user ID from the request parameters
	targetUserID := c.Param("id")
	targetUserObjectID, err := primitive.ObjectIDFromHex(targetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Extract claims to get the role of the logged-in user
	claims, _ := c.Get("user")
	role := claims.(jwt.MapClaims)["role"].(string)

	// Check permissions
	if role != "admin" && loggedInUserID != targetUserObjectID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can delete other users"})
		return
	}

	// Access the users collection
	collection := config.DB.Collection("users")

	// Find the user to ensure they exist
	var user models.User
	err = collection.FindOne(context.TODO(), bson.M{"_id": targetUserObjectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Perform the deletion
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": targetUserObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}