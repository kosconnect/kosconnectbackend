package controllers

import (
	"context"
	"net/http"
	"fmt"


	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/organisasi/kosconnectbackend/config"
	"github.com/organisasi/kosconnectbackend/models"
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

// Delete user (for the logged-in user or admin)
func DeleteUser(c *gin.Context) {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if the logged-in user is deleting their own account or an admin is deleting any user
	if userID.Hex() != c.Param("id") {
		claims, _ := c.Get("user")
		role := claims.(jwt.MapClaims)["role"].(string)
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own account"})
			return
		}
	}

	// Delete user from MongoDB
	collection := config.DB.Collection("users")
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
