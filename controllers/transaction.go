package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"

	// "go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/primitive"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/organisasi/kosconnectbackend/config"
	"github.com/organisasi/kosconnectbackend/models"
)

func CreateTransaction(c *gin.Context) {
	// Ambil query string dari request
	roomID := c.Query("room_id")
	boardingHouseID := c.Query("boarding_house_id")
	ownerID := c.Query("owner_id")
	userID := c.Query("user_id")

	// Validasi input
	if roomID == "" || boardingHouseID == "" || ownerID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters"})
		return
	}

	// Convert IDs menjadi ObjectID
	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}
	boardingHouseObjectID, err := primitive.ObjectIDFromHex(boardingHouseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid boarding house ID"})
		return
	}
	ownerObjectID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID"})
		return
	}
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Ambil data kamar
	roomCollection := config.DB.Collection("rooms")
	var room models.Room
	err = roomCollection.FindOne(context.TODO(), bson.M{"_id": roomObjectID}).Decode(&room)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Ambil data boarding house
	boardingHouseCollection := config.DB.Collection("boardinghouses")
	var boardingHouse models.BoardingHouse
	err = boardingHouseCollection.FindOne(context.TODO(), bson.M{"_id": boardingHouseObjectID}).Decode(&boardingHouse)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Boarding house not found"})
		return
	}

	// Ambil data pengguna
	userCollection := config.DB.Collection("users")
	var user models.User
	err = userCollection.FindOne(context.TODO(), bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Ambil custom facilities dari body request
	var requestBody struct {
		CustomFacilityIDs []string `json:"custom_facilities"`
		PaymentTerm       string   `json:"payment_term"`
		CheckInDate       string   `json:"check_in_date"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Validasi payment term
	paymentTerm := requestBody.PaymentTerm
	validTerms := []string{"monthly", "quarterly", "semi_annual", "yearly"}
	isValidTerm := false
	for _, term := range validTerms {
		if paymentTerm == term {
			isValidTerm = true
			break
		}
	}
	if !isValidTerm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment term"})
		return
	}

	// Validasi tanggal check-in
	checkInDate, err := time.Parse("2006-01-02", requestBody.CheckInDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid check-in date format"})
		return
	}

	// Query fasilitas custom berdasarkan ID
	var customFacilities []models.CustomFacilityInfo
	customFacilityCollection := config.DB.Collection("customFacility")

	for _, customFacilityID := range requestBody.CustomFacilityIDs {
		cfObjectID, err := primitive.ObjectIDFromHex(customFacilityID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom facility ID"})
			return
		}

		var customFacility models.CustomFacility
		err = customFacilityCollection.FindOne(context.TODO(), bson.M{"_id": cfObjectID}).Decode(&customFacility)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Custom facility not found"})
			return
		}

		customFacilities = append(customFacilities, models.CustomFacilityInfo{
			CustomFacilityID: customFacility.CustomFacilityID,
			Name:             customFacility.Name,
			Price:            customFacility.Price,
		})
	}

	// Hitung total harga custom facilities
	var facilitiesPrice float64
	for _, cf := range customFacilities {
		facilitiesPrice += cf.Price
	}

	// Hitung harga kamar berdasarkan payment term
	var roomPrice float64
	switch paymentTerm {
	case "monthly":
		roomPrice = float64(room.Price.Monthly)
	case "quarterly":
		roomPrice = float64(room.Price.Quarterly)
	case "semi_annual":
		roomPrice = float64(room.Price.SemiAnnual)
	case "yearly":
		roomPrice = float64(room.Price.Yearly)
	}

	// Hitung total transaksi
	ppn := (roomPrice + facilitiesPrice) * 0.11 // PPN 11%
	total := roomPrice + facilitiesPrice + ppn

	// Buat data transaksi
	transaction := models.Transaction{
		TransactionID:   primitive.NewObjectID(),
		TransactionCode: "TX-" + primitive.NewObjectID().Hex(),
		UserID:          userObjectID,
		OwnerID:         ownerObjectID,
		BoardingHouseID: boardingHouseObjectID,
		RoomID:          roomObjectID,
		PersonalInfo: models.PersonalInfo{
			FullName:    user.FullName,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
		},
		CustomFacilities: customFacilities,
		PaymentTerm:      paymentTerm,
		CheckInDate:      checkInDate,
		Price:            roomPrice,
		FacilitiesPrice:  facilitiesPrice,
		PPN:              ppn,
		Total:            total,
		PaymentStatus:    "pending",
		PaymentMethod:    "",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Simpan transaksi ke database
	transactionCollection := config.DB.Collection("transactions")
	_, err = transactionCollection.InsertOne(context.TODO(), transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	// Kirim response sukses
	c.JSON(http.StatusOK, gin.H{
		"message":        "Transaction created successfully",
		"transaction_id": transaction.TransactionID,
	})
}

// dipakai oleh admin dan owner
func GetAllTransactions(c *gin.Context) {
	collection := config.DB.Collection("transactions")

	// Ambil semua data transaksi dari database
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}
	defer cursor.Close(context.TODO())

	// Simpan hasilnya dalam slice
	var transactions []models.Transaction
	if err = cursor.All(context.TODO(), &transactions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse transactions"})
		return
	}

	// Kembalikan data
	c.JSON(http.StatusOK, gin.H{
		"message": "Transactions fetched successfully",
		"data":    transactions,
	})
}

// untuk dapatkan transaksi berdasarkan id
func GetTransactionByID(c *gin.Context) {
	id := c.Param("id")

	transactionID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	collection := config.DB.Collection("transactions")

	var transaction models.Transaction
	err = collection.FindOne(context.TODO(), bson.M{"_id": transactionID}).Decode(&transaction)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Kembalikan data transaksi
	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction fetched successfully",
		"data":    transaction,
	})
}

// mendapatkan data transaksi berdasarkan user yang login
// A.K.A BUAT DI HALAMAN USER YA FATH / BALQIS .-fath cantik
func GetTransactionsByUser(c *gin.Context) {
	// Ambil ID user dari JWT
	claims := c.MustGet("user").(jwt.MapClaims)
	userID, ok := claims["user_id"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user ID from token"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := config.DB.Collection("transactions")

	// Ambil semua transaksi milik user
	cursor, err := collection.Find(context.TODO(), bson.M{"user_id": userObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user transactions"})
		return
	}
	defer cursor.Close(context.TODO())

	var transactions []models.Transaction
	if err = cursor.All(context.TODO(), &transactions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user transactions"})
		return
	}

	// Kembalikan data transaksi user
	c.JSON(http.StatusOK, gin.H{
		"message": "User transactions fetched successfully",
		"data":    transactions,
	})
}

// untuk dapatkan transaksi berdasarkan user tertentu oleh ADMIN
func GetTransactionsUserByAdmin(c *gin.Context) {
	userID := c.Param("userID")

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := config.DB.Collection("transactions")

	cursor, err := collection.Find(context.TODO(), bson.M{"user_id": userObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user transactions"})
		return
	}
	defer cursor.Close(context.TODO())

	var transactions []models.Transaction
	if err = cursor.All(context.TODO(), &transactions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user transactions"})
		return
	}

	// Kembalikan data transaksi user
	c.JSON(http.StatusOK, gin.H{
		"message": "User transactions fetched successfully",
		"data":    transactions,
	})
}

// INI YANG DI PAKE DI DASHBOARD OWNER YA :* YA  JADI PERHATIKAN ENDPOINTNYA T_T
func GetTransactionsByOwner(c *gin.Context) {
	// Ambil ID owner dari JWT
	claims := c.MustGet("user").(jwt.MapClaims)
	ownerID, ok := claims["user_id"].(string)
	role, _ := claims["role"].(string)

	// Validasi role: hanya "owner" yang diizinkan
	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners can access this resource"})
		return
	}

	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse owner ID from token"})
		return
	}

	ownerObjectID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID"})
		return
	}

	collection := config.DB.Collection("transactions")

	// Ambil semua transaksi milik owner
	cursor, err := collection.Find(context.TODO(), bson.M{"owner_id": ownerObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch owner transactions"})
		return
	}
	defer cursor.Close(context.TODO())

	var transactions []models.Transaction
	if err = cursor.All(context.TODO(), &transactions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse owner transactions"})
		return
	}

	// Kembalikan data transaksi owner
	c.JSON(http.StatusOK, gin.H{
		"message": "Owner transactions fetched successfully",
		"data":    transactions,
	})
}

// untuk cari data transaksi berdasarkan owner oleh ADMIN
// kali aja butuh jadi disimpen code nya
func GetTransactionsOwnerByAdmin(c *gin.Context) {
	ownerID := c.Param("ownerID")

	ownerObjectID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID"})
		return
	}

	collection := config.DB.Collection("transactions")

	cursor, err := collection.Find(context.TODO(), bson.M{"owner_id": ownerObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch owner transactions"})
		return
	}
	defer cursor.Close(context.TODO())

	var transactions []models.Transaction
	if err = cursor.All(context.TODO(), &transactions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse owner transactions"})
		return
	}

	// Kembalikan data transaksi owner
	c.JSON(http.StatusOK, gin.H{
		"message": "Owner transactions fetched successfully",
		"data":    transactions,
	})
}

// untuk dapatkan transaksi berdasarkan status
func GetTransactionsByPaymentStatus(c *gin.Context) {
	status := c.Param("status")

	collection := config.DB.Collection("transactions")

	cursor, err := collection.Find(context.TODO(), bson.M{"payment_status": status})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions by payment status"})
		return
	}
	defer cursor.Close(context.TODO())

	var transactions []models.Transaction
	if err = cursor.All(context.TODO(), &transactions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse transactions by payment status"})
		return
	}

	// Kembalikan data transaksi berdasarkan status pembayaran
	c.JSON(http.StatusOK, gin.H{
		"message": "Transactions by payment status fetched successfully",
		"data":    transactions,
	})
}

// UPDATE STATUS DOANG
func UpdateTransactionPaymentStatus(c *gin.Context) {
	// Ambil ID transaksi dari parameter URL
	id := c.Param("id")
	transactionID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	// Ambil data dari token JWT
	claims := c.MustGet("user").(jwt.MapClaims)
	role := claims["role"].(string)

	// Hanya admin atau owner yang diizinkan
	if role != "admin" && role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update transactions"})
		return
	}

	// Parsing body request
	var updateData struct {
		Status      string `json:"status"`      // Status pembayaran (e.g., Pending, Paid, Cancelled)
		Description string `json:"description"` // Deskripsi tambahan, jika ada
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validasi input
	if updateData.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	collection := config.DB.Collection("transactions")

	// Cari dan perbarui transaksi
	filter := bson.M{"_id": transactionID}
	update := bson.M{
		"$set": bson.M{
			"status":      updateData.Status,
			"description": updateData.Description,
			"updated_at":  time.Now(),
		},
	}

	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	// Periksa apakah transaksi ditemukan dan diperbarui
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Berikan respons sukses
	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction updated successfully",
	})
}

// DELETE TRANSACTION (ONLY ADMIN)
func DeleteTransaction(c *gin.Context) {
	// Ambil ID transaksi dari parameter URL
	id := c.Param("id")
	transactionID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	// Ambil data dari token JWT
	claims := c.MustGet("user").(jwt.MapClaims)
	role := claims["role"].(string)

	// Hanya admin yang diizinkan
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete transactions"})
		return
	}

	collection := config.DB.Collection("transactions")

	// Cari dan hapus transaksi
	filter := bson.M{"_id": transactionID}
	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
		return
	}

	// Periksa apakah transaksi ditemukan dan dihapus
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Berikan respons sukses
	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction deleted successfully",
	})
}
