package fbo

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/julietrb1/offair-cli/models"
)

// RemoveFBO removes an FBO from an airport
func RemoveFBO(db *sqlx.DB, icao string) error {
	var airport models.Airport
	err := db.Get(&airport, "SELECT * FROM airports WHERE icao = ?", icao)
	if err != nil {
		return fmt.Errorf("airport with ICAO %s not found: %w", icao, err)
	}

	if !airport.HasFBO {
		return fmt.Errorf("airport %s does not have an FBO", icao)
	}

	// Remove FBO
	_, err = db.Exec("DELETE FROM fbos WHERE airport_id = ?", airport.ID)
	if err != nil {
		return fmt.Errorf("error removing FBO: %w", err)
	}

	// Update airport
	_, err = db.Exec("UPDATE airports SET has_fbo = FALSE WHERE id = ?", airport.ID)
	if err != nil {
		return fmt.Errorf("error updating airport: %w", err)
	}

	return nil
}
