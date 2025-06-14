package onair

import (
	"time"

	"github.com/google/uuid"
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

type Runway struct {
	ID                 uuid.UUID `json:"Id"`
	AirportID          uuid.UUID `json:"AirportId"`
	Name               string    `json:"Name"`
	Latitude           float64   `json:"Latitude"`
	Longitude          float64   `json:"Longitude"`
	MagneticHeading    int       `json:"MagneticHeading"`
	Length             int       `json:"Length"`
	Width              int       `json:"Width"`
	HasILS             bool      `json:"HasIls"`
	IlsFrequency       float64   `json:"IlsFrequency"`
	IlsId              string    `json:"IlsId"`
	IlsSlope           float64   `json:"IlsSlope"`
	IlsMagneticHeading int       `json:"IlsMagneticHeading"`
	ThresholdElevation int       `json:"ThresholdElevation"`
	SurfaceType        int       `json:"SurfaceType"`
	RunwayType         int       `json:"RunwayType"`
	ApproachLights     string    `json:"ApproachLights"`
	EndLights          bool      `json:"EndLights"`
	CenterLights       int       `json:"CenterLights"`
	EdgeLights         int       `json:"EdgeLights"`
}

type AirportLocation struct {
	ID              string  `json:"Id"`
	AirportID       string  `json:"AirportId"`
	Name            string  `json:"Name"`
	Latitude        float64 `json:"Latitude"`
	Longitude       float64 `json:"Longitude"`
	MagneticHeading float64 `json:"MagneticHeading"`
	Type            int     `json:"Type"`
}

type Airport struct {
	ID                                 uuid.UUID         `json:"Id"`
	ICAO                               string            `json:"ICAO"`
	HasNoRunways                       bool              `json:"HasNoRunways"`
	TimeOffsetInSec                    float64           `json:"TimeOffsetInSec"`
	LocalTimeOpenInHoursSinceMidnight  float64           `json:"LocalTimeOpenInHoursSinceMidnight"`
	LocalTimeCloseInHoursSinceMidnight float64           `json:"LocalTimeCloseInHoursSinceMidnight"`
	IATA                               string            `json:"IATA"`
	Name                               string            `json:"Name"`
	State                              string            `json:"State"`
	CountryCode                        string            `json:"CountryCode"`
	CountryName                        string            `json:"CountryName"`
	City                               string            `json:"City"`
	Latitude                           float64           `json:"Latitude"`
	Longitude                          float64           `json:"Longitude"`
	Elevation                          float64           `json:"Elevation"`
	HasLandRunway                      bool              `json:"HasLandRunway"`
	HasWaterRunway                     bool              `json:"HasWaterRunway"`
	HasHelipad                         bool              `json:"HasHelipad"`
	Size                               int               `json:"Size"`
	TransitionAltitude                 int               `json:"TransitionAltitude"`
	LastMETARDate                      string            `json:"LastMETARDate"`
	Runways                            []Runway          `json:"Runways"`
	AirportLocations                   []AirportLocation `json:"AirportLocations"`
	AirportFrequencies                 []any             `json:"AirportFrequencies"`
	IsNotInVanillaFSX                  bool              `json:"IsNotInVanillaFSX"`
	IsNotInVanillaP3D                  bool              `json:"IsNotInVanillaP3D"`
	IsNotInVanillaXPLANE               bool              `json:"IsNotInVanillaXPLANE"`
	IsNotInVanillaFS2020               bool              `json:"IsNotInVanillaFS2020"`
	IsClosed                           bool              `json:"IsClosed"`
	IsValid                            bool              `json:"IsValid"`
	MagVar                             float64           `json:"MagVar"`
	IsAddon                            bool              `json:"IsAddon"`
	Orientation                        float64           `json:"Orientation"`
	WikiUrl                            string            `json:"WikiUrl"`
	DisplaySceneryInSim                bool              `json:"DisplaySceneryInSim"`
	SceneryLatitude                    float64           `json:"SceneryLatitude"`
	SceneryLongitude                   float64           `json:"SceneryLongitude"`
	RandomSeed                         int               `json:"RandomSeed"`
	LastRandomSeedGeneration           string            `json:"LastRandomSeedGeneration"`
	IsMilitary                         bool              `json:"IsMilitary"`
	HasLights                          bool              `json:"HasLights"`
	IsBasecamp                         bool              `json:"IsBasecamp"`
	LastHangarFeesProcessDate          string            `json:"LastHangarFeesProcessDate"`
	MapSurfaceType                     int               `json:"MapSurfaceType"`
	CreationDate                       string            `json:"CreationDate"`
	IsInSimbrief                       bool              `json:"IsInSimbrief"`
	AirportSource                      int               `json:"AirportSource"`
	LastVeryShortRequestDate           string            `json:"LastVeryShortRequestDate"`
	LastSmallTripRequestDate           string            `json:"LastSmallTripRequestDate"`
	LastMediumTripRequestDate          string            `json:"LastMediumTripRequestDate"`
	LastShortHaulRequestDate           string            `json:"LastShortHaulRequestDate"`
	LastMediumHaulRequestDate          string            `json:"LastMediumHaulRequestDate"`
	LastLongHaulRequestDate            string            `json:"LastLongHaulRequestDate"`
	DisplayName                        string            `json:"DisplayName"`
	UTCTimeOpenInHoursSinceMidnight    float64           `json:"UTCTimeOpenInHoursSinceMidnight"`
	UTCTimeCloseInHoursSinceMidnight   float64           `json:"UTCTimeCloseInHoursSinceMidnight"`
}

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

type AircraftAtAirportResponse struct {
	Content []AircraftAtAirportContent `json:"Content"`
}
