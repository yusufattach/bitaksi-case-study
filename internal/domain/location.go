package domain

import (
	"time"
)

// Point represents a GeoJSON Point type
type Point struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

// DriverLocation represents a driver's location at a specific time
type DriverLocation struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	DriverID  string    `json:"driver_id" bson:"driver_id"`
	Location  Point     `json:"location" bson:"location"`
	Status    string    `json:"status" bson:"status"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

// LocationRequest represents a request to find drivers within a radius
type LocationRequest struct {
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
	Radius    float64 `json:"radius" validate:"required,min=0"`
}

// NewPoint creates a new GeoJSON Point
func NewPoint(lat, lon float64) Point {
	return Point{
		Type:        "Point",
		Coordinates: []float64{lon, lat}, // MongoDB expects [longitude, latitude]
	}
}

// GetCoordinates returns latitude and longitude from Point
func (p Point) GetCoordinates() (float64, float64) {
	return p.Coordinates[1], p.Coordinates[0] // returns [latitude, longitude]
}
