package menu

import (
	"github.com/AlecAivazis/survey/v2"
	"offair-cli/models"
)

// promptForAirportType prompts the user to select an airport type and updates the airport object
// Returns true if the user selected a type, false if they skipped
func promptForAirportType(airport *models.Airport) bool {
	var airportTypeOption string
	airportTypePrompt := &survey.Select{
		Message: "Select airport type:",
		Options: []string{
			"Aircraft Landing Area (ALA)",
			"Aerodrome (AD)",
			"Skip",
		},
	}
	survey.AskOne(airportTypePrompt, &airportTypeOption)

	// Update the airport object with the user-provided airport type
	if airportTypeOption == "Aerodrome (AD)" {
		airportType := "AD"
		airport.AirportType = &airportType
		return true
	} else if airportTypeOption == "Aircraft Landing Area (ALA)" {
		airportType := "ALA"
		airport.AirportType = &airportType
		return true
	}
	// If the user selects "Skip", leave the airport type as nil
	return false
}
