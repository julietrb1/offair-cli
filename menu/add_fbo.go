package menu

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"github.com/julietrb1/offair-cli/api"
	"github.com/julietrb1/offair-cli/fbo"
	"github.com/julietrb1/offair-cli/models"
	"strings"
)

// AddFBO adds an FBO at an airport
func AddFBO(db *sqlx.DB) {
	bold := color.New(color.Bold).SprintFunc()

	for {
		var icao string
		prompt := &survey.Input{
			Message: "Enter ICAO of the airport (blank to go back):",
		}
		survey.AskOne(prompt, &icao)

		if icao == "" {
			return
		}
		icao = strings.ToUpper(icao)

		var airport models.Airport
		err := db.Get(&airport, "SELECT * FROM airports WHERE icao = ?", icao)
		if err != nil {
			fmt.Printf("%s %s %s\n",
				color.YellowString("Airport with ICAO"),
				bold(icao),
				color.YellowString("not found. Fetching from the API..."))

			onairAPI, err := api.NewOnAirAPI()
			if err != nil {
				fmt.Printf("%s %v\n", color.RedString("Error initializing API client:"), err)
				fmt.Println(color.YellowString("Please set the ONAIR_API_KEY environment variable in your .env file."))
				continue
			}

			apiAirport, err := onairAPI.GetAirport(icao)
			if err != nil {
				fmt.Printf("%s %v\n", color.RedString("Error fetching airport from API:"), err)
				continue
			}

			dbAirport := api.AdaptAirportToDBModel(*apiAirport)
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

				if countryCode == "" {
					continue
				}

				countryCode = strings.ToUpper(countryCode)
				dbAirport.CountryCode = countryCode
			}

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

			if airportTypeOption == ADMenuLabel {
				airportType := "AD"
				dbAirport.AirportType = &airportType
			} else if airportTypeOption == ALAMenuLabel {
				airportType := "ALA"
				dbAirport.AirportType = &airportType
			}

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
				color.CyanString("fetched from API and added to database."))
		}

		if airport.AirportType == nil {
			promptForAirportType(&airport)

			_, err = db.NamedExec(`
				UPDATE airports SET
					airport_type = :airport_type
				WHERE id = :id
			`, airport)
			if err != nil {
				fmt.Printf("%s %v\n", color.RedString("Error updating airport:"), err)
			}
		}

		err = fbo.AddFBO(db, icao)
		if err != nil {
			fmt.Printf("%s %v\n", color.RedString("Error:"), err)
		} else {
			fmt.Printf("%s %s %s\n",
				color.GreenString("FBO added at"),
				bold(icao))
		}
		fmt.Println()
	}
}
