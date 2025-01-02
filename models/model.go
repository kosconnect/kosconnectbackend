package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FullName string             `bson:"fullname,omitempty" json:"fullname,omitempty"`
	Email    string             `bson:"email,omitempty" json:"email,omitempty"`
	Role     string             `bson:"role,omitempty" json:"role,omitempty"`
	Password string             `bson:"password,omitempty" json:"password,omitempty"`
}

type Category struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name string             `bson:"name,omitempty" json:"name,omitempty"`
	Slug string             `bson:"slug,omitempty" json:"slug,omitempty"`
}

type FacilityType struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name string             `bson:"name,omitempty" json:"name,omitempty"`
}

// BoardingHouse model
type BoardingHouse struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	OwnerID       primitive.ObjectID `bson:"owner_id,omitempty" json:"owner_id,omitempty"`
	CategoryID    primitive.ObjectID `bson:"category_id,omitempty" json:"category_id,omitempty"`
	Name          string             `bson:"name,omitempty" json:"name,omitempty"`
	Slug          string             `bson:"slug,omitempty" json:"slug,omitempty"`
	Address       string             `bson:"address,omitempty" json:"address,omitempty"`
	Longitude     float64            `bson:"longitude,omitempty" json:"longitude,omitempty"`
	Latitude      float64            `bson:"latitude,omitempty" json:"latitude,omitempty"`
	Description   string             `bson:"description,omitempty" json:"description,omitempty"`
	Facilities    []Facilities       `bson:"facilities,omitempty" json:"facilities,omitempty"`
	Images        []string           `bson:"images,omitempty" json:"images,omitempty"` // Array of image URLs
	Rules         string             `bson:"rules,omitempty" json:"rules,omitempty"`
	ClosestPlaces []ClosestPlace     `bson:"closest_places,omitempty" json:"closest_places,omitempty"`
}

// untuk simpan data facility umum di boarding house
type Facilities struct {
	FacilityID primitive.ObjectID `json:"facility_id,omitempty" bson:"facility_id,omitempty"`
	Name string `json:"name,omitempty" bson:"name,omitempty"`
}

// ClosestPlace model
type ClosestPlace struct {
	Name     string  `bson:"name,omitempty" json:"name,omitempty"`
	Distance float64 `bson:"distance,omitempty" json:"distance,omitempty"` // Jarak dalam satuan tertentu
	Unit     string  `bson:"unit,omitempty" json:"unit,omitempty"`         // "m" atau "km"
}

// RoomFacility model ini merupakan collection
type RoomFacility struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name string             `bson:"name,omitempty" json:"name,omitempty"`
}
// room custom facility
type CustomFacility struct {
	ID      string             `json:"id,omitempty" bson:"_id,omitempty"`
	Name    string             `json:"name,omitempty" bson:"name,omitempty"`
	Price   float64            `json:"price,omitempty" bson:"price,omitempty"`
	OwnerID primitive.ObjectID `json:"owner_id,omitempty" bson:"owner_id,omitempty"`
}

// Room model
type Room struct {
	ID               primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	BoardingHouseID  primitive.ObjectID   `bson:"boarding_house_id,omitempty" json:"boarding_house_id,omitempty"`
	RoomType         string               `bson:"room_type,omitempty" json:"room_type,omitempty"`
	Size             string               `bson:"size,omitempty" json:"size,omitempty"`
	Price            RoomPrice            `bson:"price,omitempty" json:"price,omitempty"`
	RoomFacilities   []RoomFacilities `bson:"room_facilities,omitempty" json:"room_facilities,omitempty"`
	CustomFacilities []CustomFacilities `bson:"custom_facilities,omitempty" json:"custom_facilities,omitempty"`
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

// tempat simpan roomfacility dari tiap ruangan
type RoomFacilities struct {
	RoomFacilityID primitive.ObjectID `bson:"roomfacility_id,omitempty" json:"roomfacility_id,omitempty"`
	Name string `bson:"name,omitempty" json:"name,omitempty"`
}

// tempat simpan custom facility yang ada di tiap kamar/ruangan/room
type CustomFacilities struct {
	CustomFacilitiesID primitive.ObjectID `bson:"customfacility_id,omitempty" json:"customfacility_id,omitempty"`
	Name    string             `json:"name,omitempty" bson:"name,omitempty"`
	Price   float64            `json:"price,omitempty" bson:"price,omitempty"`
}