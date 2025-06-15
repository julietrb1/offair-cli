package onair

import (
	"github.com/google/uuid"
)

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
