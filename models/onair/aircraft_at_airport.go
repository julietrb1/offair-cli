package onair

import (
	"github.com/google/uuid"
)

type AircraftTypeAtAirport struct {
	ID                         uuid.UUID     `json:"Id"`
	Hash                       string        `json:"Hash"`
	AircraftClassID            uuid.UUID     `json:"AircraftClassId"`
	AircraftClass              AircraftClass `json:"AircraftClass"`
	CreationDate               string        `json:"CreationDate"`
	LastModerationDate         string        `json:"LastModerationDate"`
	DisplayName                string        `json:"DisplayName"`
	TypeName                   string        `json:"TypeName"`
	FlightsCount               int           `json:"FlightsCount"`
	TimeBetweenOverhaul        int           `json:"TimeBetweenOverhaul"`
	HightimeAirframe           int           `json:"HightimeAirframe"`
	AirportMinSize             int           `json:"AirportMinSize"`
	EmptyWeight                int           `json:"emptyWeight"`
	MaximumGrossWeight         int           `json:"maximumGrossWeight"`
	EstimatedCruiseFF          int           `json:"estimatedCruiseFF"`
	Baseprice                  float64       `json:"Baseprice"`
	FuelTotalCapacityInGallons float64       `json:"FuelTotalCapacityInGallons"`
	EngineType                 int           `json:"engineType"`
	NumberOfEngines            int           `json:"numberOfEngines"`
	Seats                      int           `json:"seats"`
	NeedsCopilot               bool          `json:"needsCopilot"`
	FuelType                   int           `json:"fuelType"`
	MaximumCargoWeight         int           `json:"maximumCargoWeight"`
	MaximumRangeInHour         float64       `json:"maximumRangeInHour"`
	MaximumRangeInNM           float64       `json:"maximumRangeInNM"`
	DesignSpeedVS0             float64       `json:"designSpeedVS0"`
	DesignSpeedVS1             float64       `json:"designSpeedVS1"`
	DesignSpeedVC              float64       `json:"designSpeedVC"`
	IsDisabled                 bool          `json:"IsDisabled"`
	LuxeFactor                 float64       `json:"LuxeFactor"`
	GliderHasEngine            bool          `json:"GliderHasEngine"`
	StandardSeatWeight         float64       `json:"StandardSeatWeight"`
	IsFighter                  bool          `json:"IsFighter"`
	EquipmentLevel             int           `json:"EquipmentLevel"`
}

type AircraftAtAirportContent struct {
	ID             uuid.UUID             `json:"Id" gorm:"type:uuid"`
	AircraftTypeID string                `json:"AircraftTypeId"`
	AircraftType   AircraftTypeAtAirport `json:"AircraftType"`
}
