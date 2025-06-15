package menu

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"offair-cli/api"
	"offair-cli/fbo"
	"offair-cli/models"
	"strings"
)

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
