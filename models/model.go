package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	UserID         primitive.ObjectID `bson:"_id,omitempty" json:"user_id,omitempty"`
	FullName       string             `bson:"fullname,omitempty" json:"fullname,omitempty"`
	Email          string             `bson:"email,omitempty" json:"email,omitempty"`
	PhoneNumber    string             `bson:"phonenumber,omitempty" json:"phonenumber,omitempty"`
	Role           string             `bson:"role,omitempty" json:"role,omitempty"`
	Password       string             `bson:"password,omitempty" json:"password,omitempty"`
	Picture        string             `bson:"picture,omitempty" json:"picture,omitempty"`               // URL foto profil
	VerifiedEmail  bool               `bson:"verified_email,omitempty" json:"verified_email,omitempty"` // Status verifikasi email
	CreatedAt      time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`         // Waktu pembuatan
	UpdatedAt      time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`         // Waktu pembaruan
	IsRoleAssigned bool               `bson:"is_role_assigned" json:"is_role_assigned"`
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

// /////////////
type Order struct {
	OrderID               primitive.ObjectID   `bson:"_id,omitempty" json:"order_id,omitempty"`
	UserID                primitive.ObjectID   `bson:"user_id,omitempty" json:"user_id,omitempty"`
	BoardingHouseID       primitive.ObjectID   `bson:"boarding_house_id,omitempty" json:"boarding_house_id,omitempty"`
	RoomID                primitive.ObjectID   `bson:"room_id,omitempty" json:"room_id,omitempty"`
	RoomType              string               `bson:"room_type,omitempty" json:"room_type,omitempty"`
	CustomFacilities      []primitive.ObjectID `bson:"custom_facilities,omitempty" json:"custom_facilities,omitempty"`
	Name                  string               `json:"name,omitempty" bson:"name,omitempty"`
	Email                 string               `bson:"email,omitempty" json:"email,omitempty"`
	PhoneNumber           string               `bson:"phonenumber,omitempty" json:"phonenumber,omitempty"` // Nomor WhatsApp pemesan
	CheckInDate           string               `bson:"check_in_date,omitempty" json:"check_in_date,omitempty"`
	Duration              string               `bson:"duration,omitempty" json:"duration,omitempty"`
	TotalAmount           float64              `bson:"total_amount,omitempty" json:"total_amount,omitempty"`
	PaymentStatus         string               `bson:"payment_status,omitempty" json:"payment_status,omitempty"`
	PaymentVirtualAccount string               `bson:"payment_virtual_account,omitempty" json:"payment_virtual_account,omitempty"`
	PaymentURL            string               `bson:"payment_url,omitempty" json:"payment_url,omitempty"` // URL pembayaran
	OrderStatus           string               `bson:"order_status,omitempty" json:"order_status,omitempty"`
	OwnerID               primitive.ObjectID   `bson:"owner_id,omitempty" json:"owner_id,omitempty"` // Diambil untuk mendapatkan nomor WhatsApp pemilik dari collection users
	CreatedAt             string               `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt             string               `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type OrderStatus struct {
	Pending   string `json:"pending"`
	Accepted  string `json:"accepted"`
	Rejected  string `json:"rejected"`
	Completed string `json:"completed"`
}

type PaymentStatus struct {
	Pending string `json:"pending"`
	Failed  string `json:"failed"`
	Success string `json:"success"`
}
