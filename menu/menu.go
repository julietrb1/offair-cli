package menu

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"

	"offair-cli/fbo"
)

// MainMenu displays the main menu and handles user selection
func MainMenu(db *sqlx.DB) {
	for {
		var option string
		prompt := &survey.Select{
			Message: "Select an option:",
			Options: []string{
				"Airports",
				"FBOs",
				"Exit",
			},
		}
		survey.AskOne(prompt, &option)

		switch option {
		case "Airports":
			AirportsMenu(db)
		case "FBOs":
			FBOOptimiserMenu(db)
		case "Exit":
			fmt.Println("Goodbye!")
			return
		}
	}
}

// AirportsMenu displays the airports menu and handles user selection
func AirportsMenu(db *sqlx.DB) {
	for {
		var option string
		prompt := &survey.Select{
			Message: "Airports:",
			Options: []string{
				"Airport Lookup",
				"Modify Airport",
				"Back to Main Menu",
			},
		}
		survey.AskOne(prompt, &option)

		switch option {
		case "Airport Lookup":
			SearchAirportByICAO(db)
		case "Modify Airport":
			ModifyAirport(db)
		case "Back to Main Menu":
			return
		}
	}
}

// FBOOptimiserMenu displays the FBO optimiser menu and handles user selection
func FBOOptimiserMenu(db *sqlx.DB) {
	for {
		var option string
		prompt := &survey.Select{
			Message: "FBOs:",
			Options: []string{
				"List Airports with FBOs",
				"List Distances Between FBOs",
				"Find Distance Between Airports",
				"Find Optimal FBO Locations",
				"[PRESENTLY BROKEN] Find Redundant FBOs",
				"Back to Main Menu",
			},
		}
		survey.AskOne(prompt, &option)

		switch option {
		case "List Airports with FBOs":
			ListAirportsWithFBOs(db)
		case "List Distances Between FBOs":
			ListDistancesBetweenFBOs(db)
		case "Find Distance Between Airports":
			FindDistanceBetweenAirports(db)
		case "Find Optimal FBO Locations":
			FindOptimalFBOLocations(db)
		case "Find Redundant FBOs":
			FindRedundantFBOs(db)
		case "Back to Main Menu":
			return
		}
	}
}

// FBOOptions displays options for a selected FBO
func FBOOptions(db *sqlx.DB, icao string) {
	var option string
	prompt := &survey.Select{
		Message: fmt.Sprintf("FBO at %s:", icao),
		Options: []string{
			"Remove FBO",
			"Back",
		},
	}
	survey.AskOne(prompt, &option)

	if option == "Remove FBO" {
		err := fbo.RemoveFBO(db, icao)
		if err != nil {
			fmt.Printf("%s %v\n", color.RedString("Error:"), err)
		} else {
			fmt.Printf("FBO at %s removed successfully.\n", icao)
		}
	}
}

// ListDistancesBetweenFBOs lists the distances between all FBOs
func ListDistancesBetweenFBOs(db *sqlx.DB) {
	result, err := fbo.ListDistancesBetweenFBOs(db)
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Error:"), err)
		return
	}

	fmt.Println(result)
}
