package fbo

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/julietrb1/offair-cli/models"
)

// AddFBO adds an FBO at an airport
func AddFBO(db *sqlx.DB, icao string) error {
	var airport models.Airport
	err := db.Get(&airport, "SELECT * FROM airports WHERE icao = ?", icao)
	if err != nil {
		return fmt.Errorf("airport with ICAO %s not found: %w", icao, err)
	}

	if airport.HasFBO {
		return fmt.Errorf("airport %s already has an FBO", icao)
	}

	if airport.Latitude == nil || airport.Longitude == nil {
		return fmt.Errorf("airport %s does not have latitude or longitude information", icao)
	}

	// Add FBO
	_, err = db.Exec(`
		INSERT INTO fbos (airport_id, icao, name, latitude, longitude)
		VALUES (?, ?, ?, ?, ?)
	`, airport.ID, airport.ICAO, airport.Name+" FBO", *airport.Latitude, *airport.Longitude)
	if err != nil {
		return fmt.Errorf("error adding FBO: %w", err)
	}

	// Update airport
	_, err = db.Exec("UPDATE airports SET has_fbo = TRUE WHERE id = ?", airport.ID)
	if err != nil {
		return fmt.Errorf("error updating airport: %w", err)
	}

	return nil
}
