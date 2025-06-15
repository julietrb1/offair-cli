package menu

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"offair-cli/api"
	"offair-cli/models"
	"strings"
)

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

			// Prompt for airport type
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
				dbAirport.AirportType = &airportType
			} else if airportTypeOption == "Aircraft Landing Area (ALA)" {
				airportType := "ALA"
				dbAirport.AirportType = &airportType
			}
			// If the user selects "Skip", leave the airport type as nil

			// Insert or replace airport in DB
			_, err = db.NamedExec(`
				INSERT OR REPLACE INTO airports (
					id, name, icao, country_code, iata, state, country_name, city,
					latitude, longitude, elevation, size, is_military, has_lights,
					is_basecamp, map_surface_type, is_in_simbrief, display_name, has_fbo,
					airport_type
				) VALUES (
					:id, :name, :icao, :country_code, :iata, :state, :country_name, :city,
					:latitude, :longitude, :elevation, :size, :is_military, :has_lights,
					:is_basecamp, :map_surface_type, :is_in_simbrief, :display_name, :has_fbo,
					:airport_type
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

		// Check if airport exists but doesn't have an airport type
		if airport.AirportType == nil {
			// Prompt for airport type
			promptForAirportType(&airport)

			// Update the database with the new airport type
			_, err = db.NamedExec(`
				UPDATE airports SET
					airport_type = :airport_type
				WHERE id = :id
			`, airport)
			if err != nil {
				fmt.Printf("%s %v\n", color.RedString("Error updating airport:"), err)
			}
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

			// Display airport type if available
			if airport.AirportType != nil {
				fmt.Printf("%s %s\n",
					bold("Airport Type:"),
					*airport.AirportType)
			} else {
				fmt.Printf("%s %s\n",
					bold("Airport Type:"),
					color.YellowString("Not set"))
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
					"Modify Airport Type",
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
			case "Modify Airport Type":
				ModifyAirportType(db, &airport)
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

// ModifyAirportType allows the user to modify the airport type
func ModifyAirportType(db *sqlx.DB, airport *models.Airport) {
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()

	// Get current airport type
	var currentType string
	if airport.AirportType != nil {
		currentType = *airport.AirportType
	} else {
		currentType = "Not set"
	}

	// Prompt for airport type
	var airportTypeOption string
	airportTypePrompt := &survey.Select{
		Message: fmt.Sprintf("Current airport type: %s. Select new type:", currentType),
		Options: []string{
			"Aircraft Landing Area (ALA)",
			"Aerodrome (AD)",
			"Clear",
			"Cancel",
		},
	}
	survey.AskOne(airportTypePrompt, &airportTypeOption)

	// Update the airport object with the user-provided airport type
	if airportTypeOption == "Aerodrome (AD)" {
		airportType := "AD"
		airport.AirportType = &airportType
	} else if airportTypeOption == "Aircraft Landing Area (ALA)" {
		airportType := "ALA"
		airport.AirportType = &airportType
	} else if airportTypeOption == "Clear" {
		airport.AirportType = nil
	} else if airportTypeOption == "Cancel" {
		return
	}

	// Update the database
	_, err := db.NamedExec(`
		UPDATE airports SET
			airport_type = :airport_type
		WHERE id = :id
	`, airport)
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Error updating airport:"), err)
		return
	}

	if airportTypeOption == "Clear" {
		fmt.Printf("%s\n", color.GreenString("Airport type cleared successfully."))
	} else if airportTypeOption != "Cancel" {
		fmt.Printf("%s %s %s\n",
			color.GreenString("Airport type updated to"),
			bold(airportTypeOption),
			color.GreenString("successfully."))
	}
}
