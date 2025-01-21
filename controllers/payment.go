package controllers

import (
	"context"
	// "fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/organisasi/kosconnectbackend/config"
	"github.com/organisasi/kosconnectbackend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreatePayment(c *gin.Context) {
	// Ambil parameter transaction_id dari request
	transactionID := c.Param("transaction_id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction ID is required"})
		return
	}

	// Validasi ObjectID
	objectID, err := primitive.ObjectIDFromHex(transactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Transaction ID"})
		return
	}

	// Ambil data transaksi dari database
	transactionCollection := config.DB.Collection("transactions")
	var transaction models.Transaction
	err = transactionCollection.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&transaction)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Konfigurasi Midtrans Snap Request
	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  transaction.TransactionCode, // Gunakan kode unik dari transaksi
			GrossAmt: int64(transaction.Total),   // Total transaksi
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: transaction.PersonalInfo.FullName,
			Email: transaction.PersonalInfo.Email,
			Phone: transaction.PersonalInfo.PhoneNumber,
		},
	}

	// Panggil API Midtrans Snap untuk membuat transaksi pembayaran
	snapResp, snapErr := config.SnapClient.CreateTransaction(snapReq)
	if snapErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment: " + snapErr.Error()})
		return
	}

	// Pastikan respons Snap mengandung token dan RedirectURL
	if snapResp.Token == "" || snapResp.RedirectURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Snap response"})
		return
	}

	// Kirim respons dengan Redirect URL dari Midtrans
	c.JSON(http.StatusOK, gin.H{
		"message":     "Payment created successfully",
		"redirectURL": snapResp.RedirectURL,
	})
}

func PaymentNotification(c *gin.Context) {
	var notificationPayload map[string]interface{}

	// Bind payload dari request
	if err := c.ShouldBindJSON(&notificationPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification payload"})
		return
	}

	// Ambil OrderID dan Status dari payload
	orderID, ok := notificationPayload["order_id"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}
	transactionStatus := notificationPayload["transaction_status"].(string)

	// Update status pembayaran di database berdasarkan status
	transactionCollection := config.DB.Collection("transactions")
	updateFields := bson.M{
		"payment_status": transactionStatus,
		"updated_at":     time.Now(),
	}

	switch transactionStatus {
	case "settlement":
		updateFields["payment_status_detail"] = "Pembayaran berhasil"
	case "pending":
		updateFields["payment_status_detail"] = "Menunggu pembayaran"
	case "expire":
		updateFields["payment_status_detail"] = "Pembayaran telah kedaluwarsa"
	case "deny":
		updateFields["payment_status_detail"] = "Pembayaran ditolak"
	}

	_, err := transactionCollection.UpdateOne(
		context.TODO(),
		bson.M{"transaction_code": orderID},
		bson.M{"$set": updateFields},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment status updated successfully"})
}
