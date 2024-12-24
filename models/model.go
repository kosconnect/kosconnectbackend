package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FullName string             `bson:"fullname,omitempty" json:"fullname,omitempty"`
	Email    string             `bson:"email,omitempty" json:"email,omitempty"`
	Role     string             `bson:"role,omitempty" json:"role,omitempty"`
	Password string             `bson:"password,omitempty" json:"password,omitempty"`
}

type CustomFacility struct {
    ID    string  `json:"id,omitempty" bson:"_id,omitempty"`
    Name  string  `json:"name,omitempty" bson:"name,omitempty"`
    Price float64 `json:"price,omitempty" bson:"price,omitempty"`
	OwnerID primitive.ObjectID `json:"owner_id,omitempty" bson:"owner_id, omitempty"`
}

type Category struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name string             `bson:"name,omitempty" json:"name,omitempty"`
	Slug string             `bson:"slug,omitempty" json:"slug,omitempty"`
}

type BoardingHouse struct {
	ID primitive.ObjectID           `bson:"_id,omitempty" json:"id,omitempty"`
	OwnerID primitive.ObjectID      `bson:"owner_id,omitempty" json:"owner_id,omitempty"`
	CategoryID primitive.ObjectID   `bson:"category_id,omitempty" json:"category_id,omitempty"`
	Name            string          `bson:"name,omitempty" json:"name,omitempty"`
	Slug            string          `bson:"slug,omitempty" json:"slug,omitempty"`
	Address         string          `bson:"address,omitempty" json:"address,omitempty"`
	Longitude       float64         `bson:"longitude,omitempty" json:"logitude,omitempty"`
	Latitude        float64         `bson:"latitude,omitempty" json:"latitude,omitempty"` 
	Description     string          `bson:"description,omitempty" json:"description,omitempty"`
	Facilities      []FacilityType  `bson:"facilities,omitempty" json:"facilities,omitempty"`
	Images          []Image         `bson:"images,omitempty" json:"images,omitempty"`
	Rules          string           `bson:"rules,omitempty" json:"rules,omitempty"`
	ClosestPlaces []ClosestPlace    `bson:"closest_places,omitempty" json:"closestplace,omitempty"`
}

type FacilityType string

const (
        AC         FacilityType = "AC"
        WiFi       FacilityType = "WiFi"
        KamarMandi FacilityType = "KamarMandi"
        // Tambahkan fasilitas lain sesuai kebutuhan
)

type Image struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	URL             string              `bson:"url,omitempty" json:"url,omitempty"`
	Caption         string              `bson:"caption,omitempty" json:"caption,omitempty"`
	RoomID          *primitive.ObjectID `bson:"room_id,omitempty" json:"room_id,omitempty"`
	BoardingHouseID primitive.ObjectID  `bson:"boarding_house_id,omitempty" json:"boarding_house_id,omitempty`
}

type ClosestPlace struct {
	Name  string     `bson:"name,omitempty" json:"name,omitempty"`
	Distance float64 `bson:"distance,omitempty" json:"distance,omitempty"` // dalam meter
}