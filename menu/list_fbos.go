package menu

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"offair-cli/fbo"
)

// ListAirportsWithFBOs lists all airports with FBOs and provides options to add/remove FBOs
func ListAirportsWithFBOs(db *sqlx.DB) {
	for {
		airports, err := fbo.ListAirportsWithFBOs(db)
		if err != nil {
			fmt.Printf("%s %v\n", color.RedString("Error:"), err)
			return
		}

		// Create options list with airports and "Add FBO" option
		options := make([]string, 0, len(airports)+2)
		for _, a := range airports {
			options = append(options, fmt.Sprintf("%s (%s)", a.Name, a.ICAO))
		}
		options = append(options, "Add FBO")
		options = append(options, "Back")

		var selection string
		prompt := &survey.Select{
			Message: "Airports with FBOs:",
			Options: options,
		}
		survey.AskOne(prompt, &selection)

		if selection == "Back" {
			return
		} else if selection == "Add FBO" {
			AddFBO(db)
		} else {
			// Extract ICAO from selection (format: "Name (ICAO)")
			icao := selection[len(selection)-5 : len(selection)-1]
			FBOOptions(db, icao)
		}
	}
}
