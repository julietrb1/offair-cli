package menu

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"github.com/julietrb1/offair-cli/models"
	"github.com/julietrb1/offair-cli/models/onair"
	"github.com/julietrb1/onair-api-go-client/api"
	"strings"
)

// ModifyAirport allows the user to modify airport details
func ModifyAirport(db *sqlx.DB) {
	bold := color.New(color.Bold).SprintFunc()

	for {
		var icao string
		prompt := &survey.Input{
			Message: "Enter ICAO of the airport to modify (blank to go back):",
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

			dbAirport := onair.AdaptAirportToDBModel(*apiAirport)

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

			// Prompt for airport type
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
				color.GreenString("fetched and added to database."))

			airport = dbAirport
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

		for {
			fmt.Printf("%s %s %s %s\n",
				bold("Airport:"),
				color.CyanString(airport.Name),
				bold("("+airport.ICAO+")"),
				color.GreenString("in "+airport.CountryCode))

			if airport.CountryName != nil {
				fmt.Printf("%s %s\n",
					bold("Country Name:"),
					*airport.CountryName)
			} else {
				fmt.Printf("%s %s\n",
					bold("Country Name:"),
					color.YellowString("Not set"))
			}

			if airport.State != nil {
				fmt.Printf("%s %s\n",
					bold("State:"),
					*airport.State)
			} else {
				fmt.Printf("%s %s\n",
					bold("State:"),
					color.YellowString("Not set"))
			}

			if airport.City != nil {
				fmt.Printf("%s %s\n",
					bold("City:"),
					*airport.City)
			} else {
				fmt.Printf("%s %s\n",
					bold("City:"),
					color.YellowString("Not set"))
			}

			if airport.Latitude != nil && airport.Longitude != nil {
				fmt.Printf("%s %.6f, %.6f\n",
					bold("Location:"),
					*airport.Latitude,
					*airport.Longitude)
			}

			if airport.AirportType != nil {
				fmt.Printf("%s %s\n",
					bold("Airport Type:"),
					*airport.AirportType)
			} else {
				fmt.Printf("%s %s\n",
					bold("Airport Type:"),
					color.YellowString("Not set"))
			}

			fmt.Println()

			var option string
			modifyPrompt := &survey.Select{
				Message: "Select field to modify:",
				Options: []string{
					"Modify Country Code",
					"Modify State",
					"Modify Country Name",
					"Modify City",
					"Modify Airport Type",
					BackMenuLabel,
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
			case BackMenuLabel:
				// Break out of the inner loop and return to the ICAO input prompt
				break
			}

			if option == BackMenuLabel {
				break
			}
		}
	}
}

// ModifyCountryCode allows the user to modify the country code of an airport
func ModifyCountryCode(db *sqlx.DB, airport *models.Airport) {
	bold := color.New(color.Bold).SprintFunc()

	var countryCode string
	prompt := &survey.Input{
		Message: "Enter new country code (blank to cancel):",
		Default: airport.CountryCode,
	}
	survey.AskOne(prompt, &countryCode)

	if countryCode == "" {
		return
	}
	countryCode = strings.ToUpper(countryCode)
	airport.CountryCode = countryCode

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
		bold(countryCode))
}

// ModifyState allows the user to modify the state of an airport
func ModifyState(db *sqlx.DB, airport *models.Airport) {
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

	if state == "" {
		airport.State = nil
	} else {
		airport.State = &state
	}

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
		fmt.Printf("%s\n", color.GreenString("State cleared."))
	} else {
		fmt.Printf("%s %s %s\n",
			color.GreenString("State updated to"),
			bold(state))
	}
}

// ModifyCountryName allows the user to modify the country name of an airport
func ModifyCountryName(db *sqlx.DB, airport *models.Airport) {
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

	if countryName == "" {
		airport.CountryName = nil
	} else {
		airport.CountryName = &countryName
	}

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
		fmt.Printf("%s\n", color.GreenString("Country name cleared."))
	} else {
		fmt.Printf("%s %s %s\n",
			color.GreenString("Country name updated to"),
			bold(countryName))
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
		fmt.Printf("%s\n", color.GreenString("City cleared."))
	} else {
		fmt.Printf("%s %s %s\n",
			color.GreenString("City updated to"),
			bold(city))
	}
}

// ModifyAirportType allows the user to modify the airport type
func ModifyAirportType(db *sqlx.DB, airport *models.Airport) {
	bold := color.New(color.Bold).SprintFunc()

	var currentType string
	if airport.AirportType != nil {
		currentType = *airport.AirportType
	} else {
		currentType = NotSetMenuLabel
	}

	var airportTypeOption string
	airportTypePrompt := &survey.Select{
		Message: fmt.Sprintf("Current airport type: %s. Select new type:", currentType),
		Options: []string{
			ALAMenuLabel,
			ADMenuLabel,
			ClearMenuLabel,
			CancelMenuLabel,
		},
	}
	survey.AskOne(airportTypePrompt, &airportTypeOption)

	if airportTypeOption == ADMenuLabel {
		airportType := "AD"
		airport.AirportType = &airportType
	} else if airportTypeOption == ALAMenuLabel {
		airportType := "ALA"
		airport.AirportType = &airportType
	} else if airportTypeOption == ClearMenuLabel {
		airport.AirportType = nil
	} else if airportTypeOption == CancelMenuLabel {
		return
	}

	_, err := db.NamedExec(`
		UPDATE airports SET
			airport_type = :airport_type
		WHERE id = :id
	`, airport)
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Error updating airport:"), err)
		return
	}

	if airportTypeOption == ClearMenuLabel {
		fmt.Printf("%s\n", color.GreenString("Airport type cleared."))
	} else if airportTypeOption != CancelMenuLabel {
		fmt.Printf("%s %s %s\n",
			color.GreenString("Airport type updated to"),
			bold(airportTypeOption))
	}
}
