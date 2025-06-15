package menu

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"offair-cli/fbo"
	"offair-cli/models"
	"strings"
)

// FindDistanceBetweenAirports calculates and displays the distance between two airports
func FindDistanceBetweenAirports(db *sqlx.DB) {
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	// Prompt for first ICAO
	var icao1 string
	prompt1 := &survey.Input{
		Message: "Enter first ICAO (blank to go back):",
	}
	survey.AskOne(prompt1, &icao1)

	// Return if blank
	if icao1 == "" {
		return
	}

	// Convert to uppercase
	icao1 = strings.ToUpper(icao1)

	// Prompt for second ICAO
	var icao2 string
	prompt2 := &survey.Input{
		Message: "Enter second ICAO (blank to go back):",
	}
	survey.AskOne(prompt2, &icao2)

	// Return if blank
	if icao2 == "" {
		return
	}

	// Convert to uppercase
	icao2 = strings.ToUpper(icao2)

	// Check if both ICAOs are the same
	if icao1 == icao2 {
		fmt.Printf("%s %s\n", red("Error:"), "Both ICAOs are the same. Please enter different ICAOs.")
		return
	}

	// Fetch first airport
	var airport1 models.Airport
	err := db.Get(&airport1, "SELECT * FROM airports WHERE icao = ?", icao1)
	if err != nil {
		fmt.Printf("%s %s %s\n", red("Error:"), bold(icao1), "not found in the database.")
		return
	}

	// Fetch second airport
	var airport2 models.Airport
	err = db.Get(&airport2, "SELECT * FROM airports WHERE icao = ?", icao2)
	if err != nil {
		fmt.Printf("%s %s %s\n", red("Error:"), bold(icao2), "not found in the database.")
		return
	}

	// Check if both airports have latitude and longitude
	if airport1.Latitude == nil || airport1.Longitude == nil {
		fmt.Printf("%s %s %s\n", red("Error:"), bold(icao1), "does not have latitude or longitude information.")
		return
	}

	if airport2.Latitude == nil || airport2.Longitude == nil {
		fmt.Printf("%s %s %s\n", red("Error:"), bold(icao2), "does not have latitude or longitude information.")
		return
	}

	// Calculate distance
	distance := fbo.CalculateDistance(*airport1.Latitude, *airport1.Longitude, *airport2.Latitude, *airport2.Longitude)

	// Display result
	fmt.Printf("\n%s\n", bold(cyan("Distance Calculation Result:")))
	fmt.Printf("%s %s (%s) %s %s (%s)\n",
		bold("From:"),
		cyan(airport1.Name),
		bold(icao1),
		bold("To:"),
		cyan(airport2.Name),
		bold(icao2))
	fmt.Printf("%s %.2f %s\n\n", bold("Distance:"), distance, green("nm"))
}
