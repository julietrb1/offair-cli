package menu

import (
	"github.com/AlecAivazis/survey/v2"
	"offair-cli/models"
)

const (
	ADMenuLabel                       = "Aerodrome (AD)"
	ALAMenuLabel                      = "Aircraft Landing Area (ALA)"
	SkipMenuLabel                     = "Skip"
	SelectAirportTypeMenuLabel        = "Select airport type:"
	BackMenuLabel                     = "Back"
	AddFBOMenuLabel                   = "Add FBO"
	AirportsWithFBOsPrompt            = "Airports with FBOs:"
	ListAirportsWithFBOsMenuLabel     = "List Airports with FBOs"
	ListDistancesBetweenFBOsMenuLabel = "List Distances Between FBOs"
	RemoveFBOMenuLabel                = "Remove FBO"
	BackToMainMenuLabel               = "Back to Main Menu"
	ExitMessage                       = "Have fun out there, captain."
	CancelMenuLabel                   = "Cancel"
	ClearMenuLabel                    = "Clear"
	NotSetMenuLabel                   = "Not Set"
)

// promptForAirportType prompts the user to select an airport type and updates the airport object
// Returns true if the user selected a type, false if they skipped
func promptForAirportType(airport *models.Airport) bool {
	var airportTypeOption string

	airportTypePrompt := &survey.Select{
		Message: SelectAirportTypeMenuLabel,
		Options: []string{
			ALAMenuLabel,
			ADMenuLabel,
			SkipMenuLabel,
		},
	}
	survey.AskOne(airportTypePrompt, &airportTypeOption)

	// Update the airport object with the user-provided airport type
	if airportTypeOption == ADMenuLabel {
		airportType := "AD"
		airport.AirportType = &airportType
		return true
	} else if airportTypeOption == ALAMenuLabel {
		airportType := "ALA"
		airport.AirportType = &airportType
		return true
	}
	// If the user selects "Skip", leave the airport type as nil
	return false
}
