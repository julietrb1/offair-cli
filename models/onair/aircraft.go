package onair

import (
	"github.com/google/uuid"
	"time"
)

type AircraftClass struct {
	ID        uuid.UUID `json:"Id"`
	ShortName string    `json:"ShortName"`
	Name      string    `json:"Name"`
	Order     int       `json:"Order"`
}

type AircraftAddon struct {
	ID                            uuid.UUID `json:"Id"`
	Hash                          string    `json:"Hash"`
	AircraftTypeID                uuid.UUID `json:"AircraftTypeId"`
	CreationDate                  time.Time `json:"CreationDate"`
	LastModerationDate            time.Time `json:"LastModerationDate,omitempty"`
	FuelTotalCapacityInGallons    float64   `json:"FuelTotalCapacityInGallons"`
	DisplayName                   string    `json:"DisplayName"`
	TypeName                      string    `json:"TypeName"`
	AirFileName                   string    `json:"AirFileName"`
	SimulatorVersion              int       `json:"simulatorVersion"`
	EmptyWeight                   int       `json:"emptyWeight"`
	MaximumGrossWeight            int       `json:"maximumGrossWeight"`
	EstimatedCruiseFF             int       `json:"estimatedCruiseFF"`
	EngineType                    int       `json:"engineType"`
	NumberOfEngines               int       `json:"numberOfEngines"`
	FuelType                      int       `json:"fuelType"`
	DesignSpeedVS0                float64   `json:"designSpeedVS0"`
	DesignSpeedVS1                float64   `json:"designSpeedVS1"`
	DesignSpeedVC                 float64   `json:"designSpeedVC"`
	IsDisabled                    bool      `json:"IsDisabled"`
	AircraftDataSheetUrl          string    `json:"AircraftDataSheetUrl,omitempty"`
	AddonUrl                      string    `json:"AddonUrl,omitempty"`
	IsVanilla                     bool      `json:"IsVanilla"`
	CreatedByUserID               uuid.UUID `json:"CreatedByUserId"`
	TestedByUser                  bool      `json:"TestedByUser"`
	LastTestFlightDate            time.Time `json:"LastTestFlightDate,omitempty"`
	ConsolidatedDesignSpeedVC     float64   `json:"ConsolidatedDesignSpeedVC"`
	ConsolidatedEstimatedCruiseFF float64   `json:"ConsolidatedEstimatedCruiseFF"`
	EnableAutoConsolidation       bool      `json:"EnableAutoConsolidation"`
	ComputedMaxPayload            int       `json:"ComputedMaxPayload"`
	ComputedSeats                 int       `json:"ComputedSeats"`
}

type AircraftType struct {
	ID                         uuid.UUID     `json:"Id"`
	Hash                       string        `json:"Hash"`
	AircraftClassID            uuid.UUID     `json:"AircraftClassId"`
	AircraftClass              AircraftClass `json:"AircraftClass"`
	CreationDate               time.Time     `json:"CreationDate"`
	LastModerationDate         time.Time     `json:"LastModerationDate"`
	DisplayName                string        `json:"DisplayName"`
	TypeName                   string        `json:"TypeName"`
	FlightsCount               int           `json:"FlightsCount"`
	TimeBetweenOverhaul        int           `json:"TimeBetweenOverhaul"`
	HightimeAirframe           int           `json:"HightimeAirframe"`
	AirportMinSize             int           `json:"AirportMinSize"`
	EmptyWeight                int           `json:"emptyWeight"`
	MaximumGrossWeight         int           `json:"maximumGrossWeight"`
	EstimatedCruiseFF          int           `json:"estimatedCruiseFF"`
	BasePrice                  float64       `json:"Baseprice"`
	FuelTotalCapacityInGallons float64       `json:"FuelTotalCapacityInGallons"`
	EngineType                 int           `json:"engineType"`
	NumberOfEngines            int           `json:"numberOfEngines"`
	Seats                      int           `json:"seats"`
	NeedsCopilot               bool          `json:"needsCopilot"`
	FuelType                   int           `json:"fuelType"`
	MaximumCargoWeight         int           `json:"maximumCargoWeight"`
	MaximumRangeInHour         float64       `json:"maximumRangeInHour"`
	MaximumRangeInNM           float64       `json:"maximumRangeInNM"`
}
