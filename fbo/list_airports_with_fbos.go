package fbo

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/julietrb1/offair-cli/models"
)

// ListAirportsWithFBOs lists all airports with FBOs
func ListAirportsWithFBOs(db *sqlx.DB) ([]models.Airport, error) {
	var airports []models.Airport
	err := db.Select(&airports, "SELECT * FROM airports WHERE has_fbo = TRUE")
	if err != nil {
		return nil, fmt.Errorf("error fetching airports with FBOs: %w", err)
	}
	return airports, nil
}
