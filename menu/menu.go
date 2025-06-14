package menu

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"

	"offair-cli/api"
	"offair-cli/fbo"
	"offair-cli/models"
)

// MainMenu displays the main menu and handles user selection
func MainMenu(db *sqlx.DB) {
	for {
		var option string
		prompt := &survey.Select{
			Message: "Select an option:",
			Options: []string{
				"Airport Lookup",
				"FBO Optimiser",
				"Exit",
			},
		}
		survey.AskOne(prompt, &option)

		switch option {
		case "Airport Lookup":
			AirportLookupMenu(db)
		case "FBO Optimiser":
			FBOOptimiserMenu(db)
		case "Exit":
			fmt.Println("Goodbye!")
			return
		}
	}
}

// AirportLookupMenu displays the airport lookup menu and handles user selection
func AirportLookupMenu(db *sqlx.DB) {
	// Directly call SearchAirportByICAO
	SearchAirportByICAO(db)
}

// SearchAirportByICAO searches for an airport by ICAO
func SearchAirportByICAO(db *sqlx.DB) {
	for {
		var icao string
		prompt := &survey.Input{
			Message: "Enter ICAO (blank to go back):",
		}
		survey.AskOne(prompt, &icao)

		// If the user enters a blank ICAO, return to the previous menu
		if icao == "" {
			return
		}

		// Convert ICAO to uppercase
		icao = strings.ToUpper(icao)

		var airport models.Airport
		err := db.Get(&airport, "SELECT * FROM airports WHERE icao = ?", icao)
		if err != nil {
			fmt.Printf("Airport with ICAO %s not found. Fetching from the API...\n", icao)

			// Initialize API client
			onairAPI, err := api.NewOnAirAPI()
			if err != nil {
				fmt.Printf("Error initializing API client: %v\n", err)
				fmt.Println("Please set the ONAIR_API_KEY environment variable in your .env file.")
				continue
			}

			// Fetch airport from API
			apiAirport, err := onairAPI.GetAirport(icao)
			if err != nil {
				fmt.Printf("Error fetching airport from API: %v\n", err)
				continue
			}

			// Adapt airport for DB
			dbAirport := api.AdaptAirportToDBModel(*apiAirport)

			// Insert airport into DB
			_, err = db.NamedExec(`
				INSERT INTO airports (
					id, name, icao, country_code, iata, state, country_name, city,
					latitude, longitude, elevation, size, is_military, has_lights,
					is_basecamp, map_surface_type, is_in_simbrief, display_name, has_fbo
				) VALUES (
					:id, :name, :icao, :country_code, :iata, :state, :country_name, :city,
					:latitude, :longitude, :elevation, :size, :is_military, :has_lights,
					:is_basecamp, :map_surface_type, :is_in_simbrief, :display_name, :has_fbo
				)
			`, dbAirport)
			if err != nil {
				fmt.Printf("Error inserting airport into database: %v\n", err)
				continue
			}

			fmt.Printf("Airport %s (%s) fetched from API and added to database.\n", dbAirport.Name, dbAirport.ICAO)

			// Update airport variable
			airport = dbAirport
		}

		// Define color functions
		bold := color.New(color.Bold).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		fmt.Printf("%s %s %s %s\n",
			bold("Airport found:"),
			cyan(airport.Name),
			bold("("+airport.ICAO+")"),
			green("in "+airport.CountryCode))

		if airport.Latitude != nil && airport.Longitude != nil {
			fmt.Printf("%s %.6f, %.6f\n",
				bold("Location:"),
				*airport.Latitude,
				*airport.Longitude)
		}

		fboStatus := "No"
		if airport.HasFBO {
			fboStatus = green("Yes")
		} else {
			fboStatus = yellow("No")
		}

		fmt.Printf("%s %s\n", bold("Has FBO:"), fboStatus)
		fmt.Println() // Add a blank line for better readability
	}
}

// FBOOptimiserMenu displays the FBO optimiser menu and handles user selection
func FBOOptimiserMenu(db *sqlx.DB) {
	for {
		var option string
		prompt := &survey.Select{
			Message: "FBO Optimiser:",
			Options: []string{
				"List Airports with FBOs",
				"List Distances Between FBOs",
				"Find Optimal FBO Locations",
				"Back to Main Menu",
			},
		}
		survey.AskOne(prompt, &option)

		switch option {
		case "List Airports with FBOs":
			ListAirportsWithFBOs(db)
		case "List Distances Between FBOs":
			ListDistancesBetweenFBOs(db)
		case "Find Optimal FBO Locations":
			FindOptimalFBOLocations(db)
		case "Back to Main Menu":
			return
		}
	}
}

// ListAirportsWithFBOs lists all airports with FBOs and provides options to add/remove FBOs
func ListAirportsWithFBOs(db *sqlx.DB) {
	for {
		airports, err := fbo.ListAirportsWithFBOs(db)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
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
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("FBO at %s removed successfully.\n", icao)
		}
	}
}

// AddFBO adds an FBO at an airport
func AddFBO(db *sqlx.DB) {
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	for {
		var icao string
		prompt := &survey.Input{
			Message: "Enter ICAO of the airport (blank to go back):",
		}
		survey.AskOne(prompt, &icao)

		// If the user enters a blank ICAO, return to the previous menu
		if icao == "" {
			return
		}

		// Convert ICAO to uppercase
		icao = strings.ToUpper(icao)

		// Check if airport exists in database
		var airport models.Airport
		err := db.Get(&airport, "SELECT * FROM airports WHERE icao = ?", icao)
		if err != nil {
			fmt.Printf("%s %s %s\n",
				yellow("Airport with ICAO"),
				bold(icao),
				yellow("not found. Fetching from the API..."))

			// Initialize API client
			onairAPI, err := api.NewOnAirAPI()
			if err != nil {
				fmt.Printf("%s %v\n", red("Error initializing API client:"), err)
				fmt.Println(yellow("Please set the ONAIR_API_KEY environment variable in your .env file."))
				continue
			}

			// Fetch airport from API
			apiAirport, err := onairAPI.GetAirport(icao)
			if err != nil {
				fmt.Printf("%s %v\n", red("Error fetching airport from API:"), err)
				continue
			}

			// Adapt airport for DB
			dbAirport := api.AdaptAirportToDBModel(*apiAirport)

			// Insert airport into DB
			_, err = db.NamedExec(`
				INSERT INTO airports (
					id, name, icao, country_code, iata, state, country_name, city,
					latitude, longitude, elevation, size, is_military, has_lights,
					is_basecamp, map_surface_type, is_in_simbrief, display_name, has_fbo
				) VALUES (
					:id, :name, :icao, :country_code, :iata, :state, :country_name, :city,
					:latitude, :longitude, :elevation, :size, :is_military, :has_lights,
					:is_basecamp, :map_surface_type, :is_in_simbrief, :display_name, :has_fbo
				)
			`, dbAirport)
			if err != nil {
				fmt.Printf("%s %v\n", red("Error inserting airport into database:"), err)
				continue
			}

			fmt.Printf("%s %s %s %s\n",
				cyan("Airport"),
				bold(dbAirport.Name),
				bold("("+dbAirport.ICAO+")"),
				cyan("fetched from API and added to database."))
		}

		// Now try to add the FBO
		err = fbo.AddFBO(db, icao)
		if err != nil {
			fmt.Printf("%s %v\n", red("Error:"), err)
		} else {
			fmt.Printf("%s %s %s\n",
				green("FBO added at"),
				bold(icao),
				green("successfully."))
		}
		fmt.Println() // Add a blank line for better readability
	}
}

// ListDistancesBetweenFBOs lists the distances between all FBOs
func ListDistancesBetweenFBOs(db *sqlx.DB) {
	result, err := fbo.ListDistancesBetweenFBOs(db)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(result)
}

// FindOptimalFBOLocations finds optimal locations for FBOs
func FindOptimalFBOLocations(db *sqlx.DB) {
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Get values from environment variables or use hardcoded defaults
	optimalDistanceStr := os.Getenv("FBO_NM_OPTIMAL")
	if optimalDistanceStr == "" {
		optimalDistanceStr = "800" // Default value when environment variable is not set
	}
	optimalDistance, _ := strconv.ParseFloat(optimalDistanceStr, 64)

	maxDistanceStr := os.Getenv("FBO_NM_MAX")
	if maxDistanceStr == "" {
		maxDistanceStr = "1200" // Default value when environment variable is not set
	}
	maxDistance, _ := strconv.ParseFloat(maxDistanceStr, 64)

	// Get FBO_REQ_LIGHTS environment variable (default to "true")
	requireLightsStr := os.Getenv("FBO_REQ_LIGHTS")
	if requireLightsStr == "" {
		requireLightsStr = "true" // Default value when environment variable is not set
	}
	requireLights := requireLightsStr == "true"

	// Get FBO_PREFERRED_SIZE environment variable (no default)
	preferredSizeStr := os.Getenv("FBO_PREFERRED_SIZE")
	var preferredSize *int
	if preferredSizeStr != "" {
		size, err := strconv.Atoi(preferredSizeStr)
		if err == nil && size >= 0 && size <= 5 {
			preferredSize = &size
		}
	}

	result, err := fbo.FindOptimalFBOLocations(db, optimalDistance, maxDistance, requireLights, preferredSize)
	if err != nil {
		fmt.Printf("%s %v\n", red("Error:"), err)
		return
	}

	fmt.Println(bold(cyan("Calculating optimal FBO locations...")))
	fmt.Println(result)
}
