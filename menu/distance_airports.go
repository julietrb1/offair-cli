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
	bold := color.New(color.Bold).SprintFunc()

	var icao1 string
	prompt1 := &survey.Input{
		Message: "Enter first ICAO (blank to go back):",
	}
	survey.AskOne(prompt1, &icao1)

	if icao1 == "" {
		return
	}
	icao1 = strings.ToUpper(icao1)

	var icao2 string
	prompt2 := &survey.Input{
		Message: "Enter second ICAO (blank to go back):",
	}
	survey.AskOne(prompt2, &icao2)

	if icao2 == "" {
		return
	}
	icao2 = strings.ToUpper(icao2)
	if icao1 == icao2 {
		fmt.Printf("%s %s\n", color.RedString("Error:"), "Both ICAOs are the same. Please enter different ICAOs.")
		return
	}

	var airport1 models.Airport
	err := db.Get(&airport1, "SELECT * FROM airports WHERE icao = ?", icao1)
	if err != nil {
		fmt.Printf("%s %s %s\n", color.RedString("Error:"), bold(icao1), "not found in the database.")
		return
	}

	var airport2 models.Airport
	err = db.Get(&airport2, "SELECT * FROM airports WHERE icao = ?", icao2)
	if err != nil {
		fmt.Printf("%s %s %s\n", color.RedString("Error:"), bold(icao2), "not found in the database.")
		return
	}

	if airport1.Latitude == nil || airport1.Longitude == nil {
		fmt.Printf("%s %s %s\n", color.RedString("Error:"), bold(icao1), "does not have latitude or longitude information.")
		return
	}

	if airport2.Latitude == nil || airport2.Longitude == nil {
		fmt.Printf("%s %s %s\n", color.RedString("Error:"), bold(icao2), "does not have latitude or longitude information.")
		return
	}

	distance := fbo.CalculateDistance(*airport1.Latitude, *airport1.Longitude, *airport2.Latitude, *airport2.Longitude)

	fmt.Printf("\n%s\n", bold(color.CyanString("Distance Calculation Result:")))
	fmt.Printf("%s %s (%s) %s %s (%s)\n",
		bold("From:"),
		color.CyanString(airport1.Name),
		bold(icao1),
		bold("To:"),
		color.CyanString(airport2.Name),
		bold(icao2))
	fmt.Printf("%s %.2f %s\n\n", bold("Distance:"), distance, color.GreenString("nm"))
}
