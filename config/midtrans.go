package config

import (
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
	"os"
)

// Inisialisasi SnapClient dan CoreAPIClient
var (
	SnapClient    snap.Client
	CoreAPIClient coreapi.Client
)

// InitMidtransConfig menginisialisasi konfigurasi Midtrans
func InitMidtransConfig() {
	// Mengambil Server Key dari environment variable
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")

	// Pastikan serverKey tidak kosong
	if serverKey == "" {
		panic("MIDTRANS_SERVER_KEY environment variable is not set")
	}

	// Inisialisasi SnapClient dan CoreAPIClient dengan Server Key dan Environment
	SnapClient.New(serverKey, midtrans.Sandbox) // Gunakan Sandbox untuk testing
	CoreAPIClient.New(serverKey, midtrans.Sandbox)
}
