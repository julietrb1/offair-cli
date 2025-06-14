package models

import (
	"time"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID         string    `json:"id" db:"id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	ModifiedAt time.Time `json:"modified_at" db:"modified_at"`
}

// Airport represents an airport
type Airport struct {
	BaseModel
	Name           string   `json:"name" db:"name"`
	ICAO           string   `json:"icao" db:"icao"`
	IATA           *string  `json:"iata" db:"iata"`
	State          *string  `json:"state" db:"state"`
	CountryCode    string   `json:"country_code" db:"country_code"`
	CountryName    *string  `json:"country_name" db:"country_name"`
	City           *string  `json:"city" db:"city"`
	Latitude       *float64 `json:"latitude" db:"latitude"`
	Longitude      *float64 `json:"longitude" db:"longitude"`
	Elevation      *float64 `json:"elevation" db:"elevation"`
	Size           *int     `json:"size" db:"size"`
	IsMilitary     bool     `json:"is_military" db:"is_military"`
	HasLights      bool     `json:"has_lights" db:"has_lights"`
	IsBasecamp     bool     `json:"is_basecamp" db:"is_basecamp"`
	MapSurfaceType *int     `json:"map_surface_type" db:"map_surface_type"`
	IsInSimbrief   bool     `json:"is_in_simbrief" db:"is_in_simbrief"`
	DisplayName    *string  `json:"display_name" db:"display_name"`
	HasFBO         bool     `json:"has_fbo" db:"has_fbo"`
	AirportType    *string  `json:"airport_type" db:"airport_type"`
}

// FBO represents a Fixed Base Operation
type FBO struct {
	ID        int     `json:"id" db:"id"`
	AirportID string  `json:"airport_id" db:"airport_id"`
	ICAO      string  `json:"icao" db:"icao"`
	Name      string  `json:"name" db:"name"`
	Latitude  float64 `json:"latitude" db:"latitude"`
	Longitude float64 `json:"longitude" db:"longitude"`
}
