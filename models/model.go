package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	UserID            primitive.ObjectID `bson:"_id,omitempty" json:"user_id,omitempty"`
	FullName          string             `bson:"fullname,omitempty" json:"fullname,omitempty"`
	Email             string             `bson:"email,omitempty" json:"email,omitempty"`
	PhoneNumber       string             `bson:"phonenumber,omitempty" json:"phonenumber,omitempty"`
	Role              string             `bson:"role,omitempty" json:"role,omitempty"`
	Password          string             `bson:"password,omitempty" json:"password,omitempty"`
	Picture           string             `bson:"picture,omitempty" json:"picture,omitempty"`               // URL foto profil
	VerifiedEmail     bool               `bson:"verified_email,omitempty" json:"verified_email,omitempty"` // Status verifikasi email
	CreatedAt         time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`         // Waktu pembuatan
	UpdatedAt         time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`         // Waktu pembaruan
	IsRoleAssigned    bool               `bson:"is_role_assigned" json:"is_role_assigned"`
	VerificationToken string             `bson:"verification_token,omitempty" json:"verification_token,omitempty"` // Token verifikasi
}

type Category struct {
	CategoryID primitive.ObjectID `bson:"_id,omitempty" json:"category_id,omitempty"`
	Name       string             `bson:"name,omitempty" json:"name,omitempty"`
}

// BoardingHouse model
type BoardingHouse struct {
	BoardingHouseID primitive.ObjectID   `bson:"_id,omitempty" json:"boarding_house_id,omitempty"`
	OwnerID         primitive.ObjectID   `bson:"owner_id,omitempty" json:"owner_id,omitempty"`
	CategoryID      primitive.ObjectID   `bson:"category_id,omitempty" json:"category_id,omitempty"`
	Name            string               `bson:"name,omitempty" json:"name,omitempty"`
	Slug            string               `bson:"slug,omitempty" json:"slug,omitempty"`
	Address         string               `bson:"address,omitempty" json:"address,omitempty"`
	Longitude       float64              `bson:"longitude,omitempty" json:"longitude,omitempty"`
	Latitude        float64              `bson:"latitude,omitempty" json:"latitude,omitempty"`
	Description     string               `bson:"description,omitempty" json:"description,omitempty"`
	Facilities      []primitive.ObjectID `bson:"facilities_id,omitempty" json:"facilities_id,omitempty"`
	Images          []string             `bson:"images,omitempty" json:"images,omitempty"` // Array of image URLs
	Rules           string               `bson:"rules,omitempty" json:"rules,omitempty"`
}

// untuk simpan data facility umum di boarding house
type Facility struct {
	FacilityID primitive.ObjectID `bson:"_id,omitempty" json:"facility_id,omitempty"`
	Name       string             `bson:"name,omitempty" json:"name,omitempty"`
	Type       string             `bson:"type,omitempty" json:"type,omitempty"` // "room" atau "boarding_house"
}

// simpan data custom facility
type CustomFacility struct {
	CustomFacilityID primitive.ObjectID `json:"custom_facility_id,omitempty" bson:"_id,omitempty"`
	Name             string             `json:"name,omitempty" bson:"name,omitempty"`
	Price            float64            `json:"price,omitempty" bson:"price,omitempty"`
	OwnerID          primitive.ObjectID `json:"owner_id,omitempty" bson:"owner_id,omitempty"`
}

// Room model
type Room struct {
	RoomID           primitive.ObjectID   `bson:"_id,omitempty" json:"room_id,omitempty"`
	BoardingHouseID  primitive.ObjectID   `bson:"boarding_house_id,omitempty" json:"boarding_house_id,omitempty"`
	RoomType         string               `bson:"room_type,omitempty" json:"room_type,omitempty"`
	Size             string               `bson:"size,omitempty" json:"size,omitempty"`
	Price            RoomPrice            `bson:"price,omitempty" json:"price,omitempty"`
	RoomFacilities   []primitive.ObjectID `bson:"room_facilities,omitempty" json:"room_facilities,omitempty"`
	CustomFacilities []primitive.ObjectID `bson:"custom_facilities,omitempty" json:"custom_facilities,omitempty"`
	Status           string               `bson:"status,omitempty" json:"status,omitempty"`
	NumberAvailable  int                  `bson:"number_available,omitempty" json:"number_available,omitempty"`
	Images           []string             `bson:"images,omitempty" json:"images,omitempty"` // Array of image URLs
}

// RoomPrice struct
type RoomPrice struct {
	Monthly    int `bson:"monthly,omitempty" json:"monthly,omitempty"`
	Quarterly  int `bson:"quarterly,omitempty" json:"quarterly,omitempty"`     // Per 3 bulan
	SemiAnnual int `bson:"semi_annual,omitempty" json:"semi_annual,omitempty"` // Per 6 bulan
	Yearly     int `bson:"yearly,omitempty" json:"yearly,omitempty"`
}

// transaksi
type Transaction struct {
	TransactionID    primitive.ObjectID   `bson:"_id,omitempty" json:"transaction_id,omitempty"`
	TransactionCode  string               `bson:"transaction_code,omitempty" json:"transaction_code,omitempty"`
	UserID           primitive.ObjectID   `bson:"user_id,omitempty" json:"user_id,omitempty"`
	OwnerID          primitive.ObjectID   `bson:"owner_id,omitempty" json:"owner_id,omitempty"`
	BoardingHouseID  primitive.ObjectID   `bson:"boarding_house_id,omitempty" json:"boarding_house_id,omitempty"`
	RoomID           primitive.ObjectID   `bson:"room_id,omitempty" json:"room_id,omitempty"`
	PersonalInfo     PersonalInfo         `bson:"personal_info,omitempty" json:"personal_info,omitempty"`
	CustomFacilities []CustomFacilityInfo `bson:"custom_facilities,omitempty" json:"custom_facilities,omitempty"`
	PaymentTerm      string               `bson:"payment_term,omitempty" json:"payment_term,omitempty"`
	CheckInDate      time.Time            `bson:"check_in_date,omitempty" json:"check_in_date,omitempty"`
	Price            float64              `bson:"price,omitempty" json:"price,omitempty"`
	FacilitiesPrice  float64              `bson:"facilities_price,omitempty" json:"facilities_price,omitempty"`
	PPN              float64              `bson:"ppn,omitempty" json:"ppn,omitempty"`
	Total            float64              `bson:"total,omitempty" json:"total,omitempty"`
	PaymentStatus    string               `bson:"payment_status,omitempty" json:"payment_status,omitempty"`
	PaymentMethod    string               `bson:"payment_method,omitempty" json:"payment_method,omitempty"`
	CreatedAt        time.Time            `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt        time.Time            `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type PersonalInfo struct {
	FullName    string `bson:"full_name,omitempty" json:"full_name,omitempty"`
	Email       string `bson:"email,omitempty" json:"email,omitempty"`
	PhoneNumber string `bson:"phone_number,omitempty" json:"phone_number,omitempty"`
	Address     string `bson:"address,omitempty" json:"address,omitempty"`
	Gender      string `bson:"gender,omitempty" json:"gender,omitempty"`
}

type CustomFacilityInfo struct {
	CustomFacilityID primitive.ObjectID `bson:"custom_facility_id,omitempty" json:"custom_facility_id,omitempty"`
	Name             string             `bson:"name,omitempty" json:"name,omitempty"`
	Price            float64            `bson:"price,omitempty" json:"price,omitempty"`
}

// // /////////////
// type Booking struct {
// 	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`                          // ID Booking
// 	UserID           primitive.ObjectID `json:"user_id" bson:"user_id"`                           // ID akun yang melakukan checkout
// 	UserEmail        string             `json:"user_email,omitempty" bson:"user_email,omitempty"` // Email akun pemesan (opsional)
// 	Customer         CustomerInfo       `json:"customer" bson:"customer"`                         // Informasi customer (penghuni kos)
// 	BoardingHouseID  primitive.ObjectID `json:"boarding_house_id" bson:"boarding_house_id"`       // ID Kos
// 	BoardingHouse    BoardingHouseInfo  `json:"boarding_house" bson:"boarding_house"`             // Informasi Kos (Embedded)
// 	BookingItems     []BookingItem      `json:"booking_items" bson:"booking_items"`               // Daftar item yang dipesan
// 	BookingDate      time.Time          `json:"booking_date" bson:"booking_date"`                 // Waktu booking
// 	Status           string             `json:"status" bson:"status"`                             // Status booking (Pending, Accepted, Rejected)
// 	NotificationSent bool               `json:"notification_sent" bson:"notification_sent"`       // Status notifikasi
// }

// // Informasi customer (penghuni kos)
// type CustomerInfo struct {
// 	FullName string `json:"fullname" bson:"fullname"` // Nama lengkap customer
// 	Email    string `json:"email" bson:"email"`       // Email customer
// 	Phone    string `json:"phone" bson:"phone"`       // Nomor telepon customer
// }

// type BoardingHouseInfo struct {
// 	Name    string             `json:"name" bson:"name"`         // Nama Kos
// 	Address string             `json:"address" bson:"address"`   // Alamat Kos
// 	OwnerID primitive.ObjectID `json:"owner_id" bson:"owner_id"` // ID Pemilik
// }

// type BookingItem struct {
// 	RoomID           primitive.ObjectID `json:"room_id" bson:"room_id"`                                         // ID Kamar
// 	RoomType         string             `json:"room_type" bson:"room_type"`                                     // Jenis Kamar
// 	CustomFacilities []CustomFacility   `json:"custom_facilities,omitempty" bson:"custom_facilities,omitempty"` // Fasilitas tambahan
// 	NumberOfRooms    int                `json:"number_of_rooms" bson:"number_of_rooms"`                         // Jumlah kamar dipesan
// 	SelectedPrice    string             `json:"selected_price" bson:"selected_price"`                           // Pilihan harga (Monthly, Quarterly, etc.)
// 	UnitPrice        float64            `json:"unit_price" bson:"unit_price"`                                   // Harga satuan
// 	TotalPrice       float64            `json:"total_price" bson:"total_price"`                                 // Harga total untuk item ini
// }
