package api

import (
	"fmt"
	"offair-cli/models/onair"
	"time"

	"offair-cli/models"
)

// AdaptAirportToDBModel converts an API Airport object to a DB Airport object.
func AdaptAirportToDBModel(apiAirport onair.Airport) models.Airport {
	// Create a new DB model instance
	dbAirport := models.Airport{
		BaseModel: models.BaseModel{
			ID:         fmt.Sprintf("%v", apiAirport.ID),
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
		},
		Name:        apiAirport.Name,
		ICAO:        apiAirport.ICAO,
		CountryCode: apiAirport.CountryCode,
		IsMilitary:  apiAirport.IsMilitary,
		HasLights:   apiAirport.HasLights,
		IsBasecamp:  apiAirport.IsBasecamp,
		HasFBO:      false, // Default to false, will be set by the application if needed
	}

	// Handle optional fields (pointers in DB model)
	if apiAirport.IATA != "" {
		iata := apiAirport.IATA
		dbAirport.IATA = &iata
	}

	if apiAirport.State != "" {
		state := apiAirport.State
		dbAirport.State = &state
	}

	if apiAirport.CountryName != "" {
		countryName := apiAirport.CountryName
		dbAirport.CountryName = &countryName
	}

	if apiAirport.City != "" {
		city := apiAirport.City
		dbAirport.City = &city
	}

	// Handle numeric fields
	latitude := apiAirport.Latitude
	dbAirport.Latitude = &latitude

	longitude := apiAirport.Longitude
	dbAirport.Longitude = &longitude

	elevation := apiAirport.Elevation
	dbAirport.Elevation = &elevation

	size := apiAirport.Size
	dbAirport.Size = &size

	mapSurfaceType := apiAirport.MapSurfaceType
	dbAirport.MapSurfaceType = &mapSurfaceType

	if apiAirport.DisplayName != "" {
		displayName := apiAirport.DisplayName
		dbAirport.DisplayName = &displayName
	}

	dbAirport.IsInSimbrief = apiAirport.IsInSimbrief

	return dbAirport
}

// AdaptAircraftTypeToDBModel converts an API AircraftType object to a DB AircraftType object.
func AdaptAircraftTypeToDBModel(apiAircraftType models.AircraftType) models.AircraftType {
	// Ensure CreatedAt and ModifiedAt are set
	if apiAircraftType.CreatedAt.IsZero() {
		apiAircraftType.CreatedAt = time.Now()
	}
	if apiAircraftType.ModifiedAt.IsZero() {
		apiAircraftType.ModifiedAt = time.Now()
	}

	return apiAircraftType
}
