package menu

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"github.com/julietrb1/offair-cli/api"
	"github.com/julietrb1/offair-cli/models"
	"strings"
)

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

		// (dangerously) assume an Australian ICAO if only three characters are provided
		if len(icao) == 3 {
			icao = "Y" + icao
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

			if dbAirport.CountryCode == "" && icao[0] == 'Y' {
				// Infer "AU" from above assumption
				countryCode := "AU"
				dbAirport.CountryCode = countryCode
			} else if dbAirport.CountryCode == "" {
				fmt.Printf("%s\n", color.YellowString(fmt.Sprintf("Airport with ICAO %s has no country code. Please enter a country code:", icao)))

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
			promptForAirportType(&dbAirport)

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
				fmt.Printf("Error inserting airport into database: %v\n", err)
				continue
			}

			fmt.Println(color.GreenString("Added to database."))

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
