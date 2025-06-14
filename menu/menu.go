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

// ModifyAirport allows the user to modify airport details
func ModifyAirport(db *sqlx.DB) {
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	for {
		var icao string
		prompt := &survey.Input{
			Message: "Enter ICAO of the airport to modify (blank to go back):",
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
				color.YellowString("Airport with ICAO"),
				bold(icao),
				color.YellowString("not found. Fetching from the API..."))

			// Initialize API client
			onairAPI, err := api.NewOnAirAPI()
			if err != nil {
				fmt.Printf("%s %v\n", color.RedString("Error initializing API client:"), err)
				fmt.Println(color.YellowString("Please set the ONAIR_API_KEY environment variable in your .env file."))
				continue
			}

			// Fetch airport from API
			apiAirport, err := onairAPI.GetAirport(icao)
			if err != nil {
				fmt.Printf("%s %v\n", color.RedString("Error fetching airport from API:"), err)
				continue
			}

			// Adapt airport for DB
			dbAirport := api.AdaptAirportToDBModel(*apiAirport)

			// Check if country code is empty
			if dbAirport.CountryCode == "" {
				fmt.Printf("%s %s %s\n",
					color.YellowString("Airport with ICAO"),
					bold(icao),
					color.YellowString("has no country code. Please enter a country code:"))

				var countryCode string
				countryPrompt := &survey.Input{
					Message: "Enter country code (blank to go back to ICAO input):",
				}
				survey.AskOne(countryPrompt, &countryCode)

				// If the user enters a blank country code, return to the ICAO input prompt
				if countryCode == "" {
					continue
				}

				// Convert country code to uppercase
				countryCode = strings.ToUpper(countryCode)

				// Update the airport object with the user-provided country code
				dbAirport.CountryCode = countryCode
			}

			// Insert or replace airport in DB
			_, err = db.NamedExec(`
				INSERT OR REPLACE INTO airports (
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
				fmt.Printf("%s %v\n", color.RedString("Error inserting airport into database:"), err)
				continue
			}

			fmt.Printf("%s %s\n",
				dbAirport.ICAO,
				cyan("fetched from API and added to database."))

			// Update airport variable
			airport = dbAirport
		}

		// Inner loop for modifying fields
		for {
			// Display airport information
			fmt.Printf("%s %s %s %s\n",
				bold("Airport:"),
				cyan(airport.Name),
				bold("("+airport.ICAO+")"),
				color.GreenString("in "+airport.CountryCode))

			// Display country name if available
			if airport.CountryName != nil {
				fmt.Printf("%s %s\n",
					bold("Country Name:"),
					*airport.CountryName)
			} else {
				fmt.Printf("%s %s\n",
					bold("Country Name:"),
					color.YellowString("Not set"))
			}

			// Display state if available
			if airport.State != nil {
				fmt.Printf("%s %s\n",
					bold("State:"),
					*airport.State)
			} else {
				fmt.Printf("%s %s\n",
					bold("State:"),
					color.YellowString("Not set"))
			}

			// Display city if available
			if airport.City != nil {
				fmt.Printf("%s %s\n",
					bold("City:"),
					*airport.City)
			} else {
				fmt.Printf("%s %s\n",
					bold("City:"),
					color.YellowString("Not set"))
			}

			// Display location if available
			if airport.Latitude != nil && airport.Longitude != nil {
				fmt.Printf("%s %.6f, %.6f\n",
					bold("Location:"),
					*airport.Latitude,
					*airport.Longitude)
			}

			// Removed "Has FBO" as per requirements
			fmt.Println() // Add a blank line for better readability

			// Show modification options
			var option string
			modifyPrompt := &survey.Select{
				Message: "Select field to modify:",
				Options: []string{
					"Modify Country Code",
					"Modify State",
					"Modify Country Name",
					"Modify City",
					"Back",
				},
			}
			survey.AskOne(modifyPrompt, &option)

			switch option {
			case "Modify Country Code":
				ModifyCountryCode(db, &airport)
			case "Modify State":
				ModifyState(db, &airport)
			case "Modify Country Name":
				ModifyCountryName(db, &airport)
			case "Modify City":
				ModifyCity(db, &airport)
			case "Back":
				// Break out of the inner loop and return to the ICAO input prompt
				break
			}

			// If the user selected "Back", break out of the inner loop
			if option == "Back" {
				break
			}
		}
	}
}

// ModifyCountryCode allows the user to modify the country code of an airport
func ModifyCountryCode(db *sqlx.DB, airport *models.Airport) {
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()

	var countryCode string
	prompt := &survey.Input{
		Message: "Enter new country code (blank to cancel):",
		Default: airport.CountryCode,
	}
	survey.AskOne(prompt, &countryCode)

	// If the user enters a blank country code, return
	if countryCode == "" {
		return
	}

	// Convert country code to uppercase
	countryCode = strings.ToUpper(countryCode)

	// Update the airport object
	airport.CountryCode = countryCode

	// Update the database
	_, err := db.NamedExec(`
		UPDATE airports SET
			country_code = :country_code
		WHERE id = :id
	`, airport)
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Error updating airport:"), err)
		return
	}

	fmt.Printf("%s %s %s\n",
		color.GreenString("Country code updated to"),
		bold(countryCode),
		color.GreenString("successfully."))
}

// ModifyState allows the user to modify the state of an airport
func ModifyState(db *sqlx.DB, airport *models.Airport) {
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()

	var state string
	defaultState := ""
	if airport.State != nil {
		defaultState = *airport.State
	}

	prompt := &survey.Input{
		Message: "Enter new state (blank to clear):",
		Default: defaultState,
	}
	survey.AskOne(prompt, &state)

	// Update the airport object
	if state == "" {
		airport.State = nil
	} else {
		airport.State = &state
	}

	// Update the database
	_, err := db.NamedExec(`
		UPDATE airports SET
			state = :state
		WHERE id = :id
	`, airport)
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Error updating airport:"), err)
		return
	}

	if state == "" {
		fmt.Printf("%s\n", color.GreenString("State cleared successfully."))
	} else {
		fmt.Printf("%s %s %s\n",
			color.GreenString("State updated to"),
			bold(state),
			color.GreenString("successfully."))
	}
}

// ModifyCountryName allows the user to modify the country name of an airport
func ModifyCountryName(db *sqlx.DB, airport *models.Airport) {
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()

	var countryName string
	defaultCountryName := ""
	if airport.CountryName != nil {
		defaultCountryName = *airport.CountryName
	}

	prompt := &survey.Input{
		Message: "Enter new country name (blank to clear):",
		Default: defaultCountryName,
	}
	survey.AskOne(prompt, &countryName)

	// Update the airport object
	if countryName == "" {
		airport.CountryName = nil
	} else {
		airport.CountryName = &countryName
	}

	// Update the database
	_, err := db.NamedExec(`
		UPDATE airports SET
			country_name = :country_name
		WHERE id = :id
	`, airport)
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Error updating airport:"), err)
		return
	}

	if countryName == "" {
		fmt.Printf("%s\n", color.GreenString("Country name cleared successfully."))
	} else {
		fmt.Printf("%s %s %s\n",
			color.GreenString("Country name updated to"),
			bold(countryName),
			color.GreenString("successfully."))
	}
}

// ModifyCity allows the user to modify the city of an airport
func ModifyCity(db *sqlx.DB, airport *models.Airport) {
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()

	var city string
	defaultCity := ""
	if airport.City != nil {
		defaultCity = *airport.City
	}

	prompt := &survey.Input{
		Message: "Enter new city (blank to clear):",
		Default: defaultCity,
	}
	survey.AskOne(prompt, &city)

	// Update the airport object
	if city == "" {
		airport.City = nil
	} else {
		airport.City = &city
	}

	// Update the database
	_, err := db.NamedExec(`
		UPDATE airports SET
			city = :city
		WHERE id = :id
	`, airport)
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Error updating airport:"), err)
		return
	}

	if city == "" {
		fmt.Printf("%s\n", color.GreenString("City cleared successfully."))
	} else {
		fmt.Printf("%s %s %s\n",
			color.GreenString("City updated to"),
			bold(city),
			color.GreenString("successfully."))
	}
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
				// Define color functions
				fmt.Printf("%s %v\n", color.RedString("Error initializing API client:"), err)
				fmt.Println("Please set the ONAIR_API_KEY environment variable in your .env file.")
				continue
			}

			// Fetch airport from API
			apiAirport, err := onairAPI.GetAirport(icao)
			if err != nil {
				// Define color functions
				redBold := color.New(color.FgRed, color.Bold).SprintFunc()
				fmt.Printf("%s %v\n", redBold("Error fetching airport from API:"), err)
				continue
			}

			// Adapt airport for DB
			dbAirport := api.AdaptAirportToDBModel(*apiAirport)

			// Check if country code is empty
			if dbAirport.CountryCode == "" {
				// Define color functions
				orange := color.New(color.FgYellow).Add(color.Bold).SprintFunc()
				fmt.Printf("%s\n", orange(fmt.Sprintf("Airport with ICAO %s has no country code. Please enter a country code:", icao)))

				var countryCode string
				countryPrompt := &survey.Input{
					Message: "Enter country code (blank to go back to ICAO input):",
				}
				survey.AskOne(countryPrompt, &countryCode)

				// If the user enters a blank country code, return to the ICAO input prompt
				if countryCode == "" {
					continue
				}

				// Convert country code to uppercase
				countryCode = strings.ToUpper(countryCode)

				// Update the airport object with the user-provided country code
				dbAirport.CountryCode = countryCode
			}

			// Insert or replace airport in DB
			_, err = db.NamedExec(`
				INSERT OR REPLACE INTO airports (
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

			fmt.Println(color.GreenString("Added to database."))

			// Update airport variable
			airport = dbAirport
		}

		// Define color functions
		bold := color.New(color.Bold).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()

		fmt.Printf("%s %s %s %s\n",
			bold("Airport found:"),
			cyan(airport.Name),
			bold("("+airport.ICAO+")"),
			color.GreenString("in "+airport.CountryCode))

		if airport.Latitude != nil && airport.Longitude != nil {
			fmt.Printf("%s %.6f, %.6f\n",
				bold("Location:"),
				*airport.Latitude,
				*airport.Longitude)
		}

		fmt.Println()
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

// AddFBO adds an FBO at an airport
func AddFBO(db *sqlx.DB) {
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

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
				color.YellowString("Airport with ICAO"),
				bold(icao),
				color.YellowString("not found. Fetching from the API..."))

			// Initialize API client
			onairAPI, err := api.NewOnAirAPI()
			if err != nil {
				fmt.Printf("%s %v\n", color.RedString("Error initializing API client:"), err)
				fmt.Println(color.YellowString("Please set the ONAIR_API_KEY environment variable in your .env file."))
				continue
			}

			// Fetch airport from API
			apiAirport, err := onairAPI.GetAirport(icao)
			if err != nil {
				fmt.Printf("%s %v\n", color.RedString("Error fetching airport from API:"), err)
				continue
			}

			// Adapt airport for DB
			dbAirport := api.AdaptAirportToDBModel(*apiAirport)

			// Check if country code is empty
			if dbAirport.CountryCode == "" {
				fmt.Printf("%s %s %s\n",
					color.YellowString("Airport with ICAO"),
					bold(icao),
					color.YellowString("has no country code. Please enter a country code:"))

				var countryCode string
				countryPrompt := &survey.Input{
					Message: "Enter country code (blank to go back to ICAO input):",
				}
				survey.AskOne(countryPrompt, &countryCode)

				// If the user enters a blank country code, return to the ICAO input prompt
				if countryCode == "" {
					continue
				}

				// Convert country code to uppercase
				countryCode = strings.ToUpper(countryCode)

				// Update the airport object with the user-provided country code
				dbAirport.CountryCode = countryCode
			}

			// Insert or replace airport in DB
			_, err = db.NamedExec(`
				INSERT OR REPLACE INTO airports (
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
				fmt.Printf("%s %v\n", color.RedString("Error inserting airport into database:"), err)
				continue
			}

			fmt.Printf("%s %s\n",
				dbAirport.ICAO,
				cyan("fetched from API and added to database."))
		}

		// Now try to add the FBO
		err = fbo.AddFBO(db, icao)
		if err != nil {
			fmt.Printf("%s %v\n", color.RedString("Error:"), err)
		} else {
			fmt.Printf("%s %s %s\n",
				color.GreenString("FBO added at"),
				bold(icao),
				color.GreenString("successfully."))
		}
		fmt.Println() // Add a blank line for better readability
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

// FindOptimalFBOLocations finds optimal locations for FBOs
func FindOptimalFBOLocations(db *sqlx.DB) {
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

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
		fmt.Printf("%s %v\n", color.RedString("Error:"), err)
		return
	}

	fmt.Println(bold(cyan("Calculating optimal FBO locations...")))
	fmt.Println(result)
}
