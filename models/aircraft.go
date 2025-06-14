package models

import (
	"time"
)

// AircraftType represents an aircraft type
type AircraftType struct {
	BaseModel
	Hash                       string    `json:"hash" db:"hash"`
	AircraftClassID            string    `json:"aircraft_class_id" db:"aircraft_class_id"`
	CreationDate               time.Time `json:"creation_date" db:"creation_date"`
	LastModerationDate         time.Time `json:"last_moderation_date" db:"last_moderation_date"`
	DisplayName                string    `json:"display_name" db:"display_name"`
	TypeName                   string    `json:"type_name" db:"type_name"`
	FlightsCount               int       `json:"flights_count" db:"flights_count"`
	TimeBetweenOverhaul        int       `json:"time_between_overhaul" db:"time_between_overhaul"`
	HightimeAirframe           int       `json:"hightime_airframe" db:"hightime_airframe"`
	AirportMinSize             int       `json:"airport_min_size" db:"airport_min_size"`
	EmptyWeight                float64   `json:"empty_weight" db:"empty_weight"`
	MaximumGrossWeight         float64   `json:"maximum_gross_weight" db:"maximum_gross_weight"`
	EstimatedCruiseFF          float64   `json:"estimated_cruise_ff" db:"estimated_cruise_ff"`
	BasePrice                  float64   `json:"base_price" db:"base_price"`
	FuelTotalCapacityInGallons float64   `json:"fuel_total_capacity_in_gallons" db:"fuel_total_capacity_in_gallons"`
	EngineType                 int       `json:"engine_type" db:"engine_type"`
	NumberOfEngines            int       `json:"number_of_engines" db:"number_of_engines"`
	Seats                      int       `json:"seats" db:"seats"`
	NeedsCopilot               bool      `json:"needs_copilot" db:"needs_copilot"`
	FuelType                   int       `json:"fuel_type" db:"fuel_type"`
	MaximumCargoWeight         float64   `json:"maximum_cargo_weight" db:"maximum_cargo_weight"`
	MaximumRangeInHour         float64   `json:"maximum_range_in_hour" db:"maximum_range_in_hour"`
	MaximumRangeInNM           float64   `json:"maximum_range_in_nm" db:"maximum_range_in_nm"`
	DesignSpeedVS0             float64   `json:"design_speed_vs0" db:"design_speed_vs0"`
	DesignSpeedVS1             float64   `json:"design_speed_vs1" db:"design_speed_vs1"`
	DesignSpeedVC              float64   `json:"design_speed_vc" db:"design_speed_vc"`
	IsDisabled                 bool      `json:"is_disabled" db:"is_disabled"`
	LuxeFactor                 float64   `json:"luxe_factor" db:"luxe_factor"`
	GliderHasEngine            bool      `json:"glider_has_engine" db:"glider_has_engine"`
	StandardSeatWeight         float64   `json:"standard_seat_weight" db:"standard_seat_weight"`
	IsFighter                  bool      `json:"is_fighter" db:"is_fighter"`
	EquipmentLevel             int       `json:"equipment_level" db:"equipment_level"`
}

// AircraftTypeAtAirport represents an aircraft type at an airport
type AircraftTypeAtAirport struct {
	AircraftType
	Count int `json:"count" db:"count"`
}

// AircraftAtAirportResponse represents the response from the API for aircraft at an airport
type AircraftAtAirportResponse struct {
	Content []AircraftTypeAtAirport `json:"Content"`
}
