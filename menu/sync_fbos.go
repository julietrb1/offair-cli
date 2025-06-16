package menu

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"

	"github.com/julietrb1/offair-cli/api"
	"github.com/julietrb1/offair-cli/models"
)

func SyncFBOs(db *sqlx.DB) {
	// Get company ID from environment variable
	companyID := os.Getenv("ONAIR_COMPANY_ID")
	if companyID == "" {
		fmt.Printf("%s %s\n", color.RedString("Error:"), "ONAIR_COMPANY_ID is not set in the environment")
		return
	}

	// Create OnAir API client
	onairAPI, err := api.NewOnAirAPI()
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Error:"), err)
		return
	}

	// Get FBOs from OnAir API
	fmt.Println("Fetching FBOs from OnAir API...")
	fbos, err := onairAPI.GetCompanyFBOs(companyID)
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Error:"), err)
		return
	}

	if len(fbos) == 0 {
		fmt.Println("No FBOs found for the company.")
		return
	}

	fmt.Printf("Found %d FBOs from API.\n", len(fbos))

	// Begin transaction
	tx, err := db.Beginx()
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Error:"), err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get existing FBOs from database
	var existingFBOs []models.FBO
	err = tx.Select(&existingFBOs, "SELECT * FROM fbos")
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Error fetching existing FBOs:"), err)
		return
	}

	fmt.Printf("Found %d existing FBOs in database.\n", len(existingFBOs))

	// Create maps for easier lookup
	existingFBOMap := make(map[string]models.FBO)
	for _, fbo := range existingFBOs {
		existingFBOMap[fbo.AirportID] = fbo
	}

	// Track which airports have FBOs
	airportsWithFBOs := make(map[string]bool)

	// Process FBOs from API
	var added, updated, unchanged int
	for _, fbo := range fbos {
		// Convert OnAir FBO to DB model
		dbFBO := api.AdaptFBOToDBModel(fbo)
		airportsWithFBOs[dbFBO.AirportID] = true

		// Check if FBO already exists
		existingFBO, exists := existingFBOMap[dbFBO.AirportID]
		if !exists {
			// FBO doesn't exist, add it
			_, err = tx.Exec(`
				INSERT INTO fbos (airport_id, icao, name, latitude, longitude)
				VALUES (?, ?, ?, ?, ?)
			`, dbFBO.AirportID, dbFBO.ICAO, dbFBO.Name, dbFBO.Latitude, dbFBO.Longitude)
			if err != nil {
				fmt.Printf("%s %v\n", color.RedString("Error inserting FBO:"), err)
				continue
			}

			// Update airport
			_, err = tx.Exec("UPDATE airports SET has_fbo = TRUE WHERE id = ?", dbFBO.AirportID)
			if err != nil {
				fmt.Printf("%s %v\n", color.RedString("Error updating airport:"), err)
				continue
			}

			fmt.Printf("Added FBO at %s (%s)\n", dbFBO.Name, dbFBO.ICAO)
			added++
		} else {
			// FBO exists, check if it needs to be updated
			if existingFBO.Name != dbFBO.Name ||
				existingFBO.Latitude != dbFBO.Latitude ||
				existingFBO.Longitude != dbFBO.Longitude {
				// Update FBO
				_, err = tx.Exec(`
					UPDATE fbos 
					SET name = ?, latitude = ?, longitude = ?
					WHERE airport_id = ?
				`, dbFBO.Name, dbFBO.Latitude, dbFBO.Longitude, dbFBO.AirportID)
				if err != nil {
					fmt.Printf("%s %v\n", color.RedString("Error updating FBO:"), err)
					continue
				}

				fmt.Printf("Updated FBO at %s (%s)\n", dbFBO.Name, dbFBO.ICAO)
				updated++
			} else {
				// FBO is unchanged
				unchanged++
			}
		}
	}

	// Remove FBOs that exist in the database but not in the API
	var removed int
	for _, fbo := range existingFBOs {
		if !airportsWithFBOs[fbo.AirportID] {
			// Remove FBO
			_, err = tx.Exec("DELETE FROM fbos WHERE airport_id = ?", fbo.AirportID)
			if err != nil {
				fmt.Printf("%s %v\n", color.RedString("Error removing FBO:"), err)
				continue
			}

			// Update airport
			_, err = tx.Exec("UPDATE airports SET has_fbo = FALSE WHERE id = ?", fbo.AirportID)
			if err != nil {
				fmt.Printf("%s %v\n", color.RedString("Error updating airport:"), err)
				continue
			}

			fmt.Printf("Removed FBO at %s (%s)\n", fbo.Name, fbo.ICAO)
			removed++
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Error:"), err)
		return
	}

	fmt.Printf("%s FBOs synchronized successfully.\n", color.GreenString("Success:"))
	fmt.Printf("  Added: %d\n", added)
	fmt.Printf("  Updated: %d\n", updated)
	fmt.Printf("  Unchanged: %d\n", unchanged)
	fmt.Printf("  Removed: %d\n", removed)
	fmt.Printf("  Total: %d\n", added+updated+unchanged)
}
