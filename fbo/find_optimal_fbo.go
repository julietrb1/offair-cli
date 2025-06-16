package fbo

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"github.com/julietrb1/offair-cli/models"
	"math"
	"strings"
)

// FindOptimalFBOLocations finds optimal locations for FBOs
func FindOptimalFBOLocations(db *sqlx.DB, optimalDistance, maxDistance float64, requireLights bool, preferredSize *int) (string, error) {
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	// Get all airports
	var airports []models.Airport
	err := db.Select(&airports, "SELECT * FROM airports WHERE latitude IS NOT NULL AND longitude IS NOT NULL")
	if err != nil {
		return "", fmt.Errorf("error fetching airports: %w", err)
	}

	// Get existing FBOs
	var existingFBOs []models.Airport
	err = db.Select(&existingFBOs, "SELECT * FROM airports WHERE has_fbo = TRUE")
	if err != nil {
		return "", fmt.Errorf("error fetching existing FBOs: %w", err)
	}

	// Check if we have enough FBOs with valid coordinates
	if len(existingFBOs) < 2 {
		return bold(yellow("There are fewer than 2 FBOs in the network. No optimization analysis possible.")), nil
	}

	// Count FBOs with valid coordinates
	validFBOCount := 0
	for _, fbo := range existingFBOs {
		if fbo.Latitude != nil && fbo.Longitude != nil {
			validFBOCount++
		}
	}

	if validFBOCount < 2 {
		return bold(yellow(fmt.Sprintf(
			"Found %d FBOs in total, but only %d have valid latitude/longitude information. "+
				"At least 2 FBOs with coordinates are needed for optimization analysis.",
			len(existingFBOs), validFBOCount))), nil
	}

	// Filter out airports that already have FBOs and apply other filters
	var candidateAirports []models.Airport
	for _, airport := range airports {
		// Skip airports that already have FBOs
		hasExistingFBO := false
		for _, fbo := range existingFBOs {
			if airport.ID == fbo.ID {
				hasExistingFBO = true
				break
			}
		}
		if hasExistingFBO {
			continue
		}

		// Skip airports without lights if required
		if requireLights && !airport.HasLights {
			continue
		}

		candidateAirports = append(candidateAirports, airport)
	}

	// Calculate scores for each candidate airport
	type AirportScore struct {
		Airport models.Airport
		Score   float64
	}

	var airportScores []AirportScore
	for _, candidate := range candidateAirports {
		// Skip airports without latitude/longitude
		if candidate.Latitude == nil || candidate.Longitude == nil {
			continue
		}

		// Calculate distances to existing FBOs
		var distances []float64
		for _, fbo := range existingFBOs {
			// Skip FBOs without latitude/longitude
			if fbo.Latitude == nil || fbo.Longitude == nil {
				continue
			}

			distance := CalculateDistance(*candidate.Latitude, *candidate.Longitude, *fbo.Latitude, *fbo.Longitude)
			distances = append(distances, distance)
		}

		// Skip if no distances were calculated
		if len(distances) == 0 {
			continue
		}

		// Count how many connections are within the optimal range
		optimalConnections := 0
		totalConnections := len(distances)

		// Calculate a score based on how many connections are within the optimal range
		// Higher score is better (100 is perfect)
		score := 0.0

		for _, distance := range distances {
			// Calculate how close this connection is to the optimal distance (as a percentage)
			// 100% means exactly at optimal distance, 0% means very far from optimal
			connectionScore := 100.0 - math.Min(100.0, (math.Abs(distance-optimalDistance)/optimalDistance)*100.0)

			// If the connection is within 20% of the optimal distance, count it as an optimal connection
			if math.Abs(distance-optimalDistance) <= 0.2*optimalDistance {
				optimalConnections++
			}

			// Add this connection's score to the total
			score += connectionScore
		}

		// If there are no eligible connections, set the score to zero
		if optimalConnections == 0 {
			score = 0.0
		} else {
			// Average the scores across all connections
			score = score / float64(totalConnections)

			// Bonus for having many connections within optimal range
			optimalRatio := float64(optimalConnections) / float64(totalConnections)
			score += optimalRatio * 20.0 // Up to 20 bonus points for having all connections optimal
		}

		// Cap at 100
		if score > 100.0 {
			score = 100.0
		}

		// Apply size preference if specified
		if preferredSize != nil && candidate.Size != nil {
			size := *candidate.Size
			preferredSizeVal := *preferredSize

			if size == preferredSizeVal {
				// Moderate positive weight for exact match
				score += 15.0
			} else if size == preferredSizeVal+1 || size == preferredSizeVal-1 {
				// Lesser positive weight for one size above or below
				score += 7.5
			}
		}

		// Apply negative weight for airports without lights if not required
		if !requireLights && !candidate.HasLights {
			score -= 10.0
		}

		// Ensure score is not negative
		if score < 0 {
			score = 0
		}

		// Cap at 100
		if score > 100.0 {
			score = 100.0
		}

		// Round down to nearest whole number
		score = math.Floor(score)

		airportScores = append(airportScores, AirportScore{
			Airport: candidate,
			Score:   score,
		})
	}

	// Filter out airports with zero scores
	var nonZeroScores []AirportScore
	for _, as := range airportScores {
		if as.Score > 0 {
			nonZeroScores = append(nonZeroScores, as)
		}
	}
	airportScores = nonZeroScores

	// Sort airports by score (higher is better)
	for i := 0; i < len(airportScores); i++ {
		for j := i + 1; j < len(airportScores); j++ {
			if airportScores[i].Score < airportScores[j].Score {
				airportScores[i], airportScores[j] = airportScores[j], airportScores[i]
			}
		}
	}

	// Build result string
	result := fmt.Sprintf("%s %s %.2f nm, %s %.2f nm\n",
		bold("Using:"),
		bold("optimal distance:"), optimalDistance,
		bold("maximum distance:"), maxDistance)

	// Add information about lights requirement
	if requireLights {
		result += fmt.Sprintf("%s %s\n",
			bold("Requiring airports with lights:"),
			green("Yes"))
	} else {
		result += fmt.Sprintf("%s %s %s\n",
			bold("Requiring airports with lights:"),
			yellow("No"),
			yellow("(airports without lights receive a score penalty)"))
	}

	// Add information about preferred size if specified
	if preferredSize != nil {
		result += fmt.Sprintf("%s %s %s %s\n",
			bold("Preferred airport size:"),
			green(fmt.Sprintf("%d", *preferredSize)),
			green("(exact match receives bonus points,"),
			green("sizes within ±1 receive smaller bonus)"))
	}

	result += fmt.Sprintf("%s %d airports, %d existing FBOs, and %d candidate airports.\n",
		bold("Found:"),
		len(airports), len(existingFBOs), len(candidateAirports))

	// Show top 10 recommended airports
	result += fmt.Sprintf("\n%s\n", bold(cyan("Top recommended airports for new FBOs:")))
	limit := 10
	if len(airportScores) < limit {
		limit = len(airportScores)
	}

	// Use color functions for scores

	for i := 0; i < limit; i++ {
		airport := airportScores[i].Airport
		score := airportScores[i].Score

		// Track connections and their contribution to score
		type Connection struct {
			ICAO         string
			Distance     float64
			Contribution float64
		}

		var connections []Connection
		totalConnections := 0

		for _, fbo := range existingFBOs {
			if fbo.Latitude == nil || fbo.Longitude == nil || airport.Latitude == nil || airport.Longitude == nil {
				continue
			}

			distance := CalculateDistance(*airport.Latitude, *airport.Longitude, *fbo.Latitude, *fbo.Longitude)
			totalConnections++

			// Calculate how much this connection contributes to the score
			contribution := 100.0 - math.Min(100.0, (math.Abs(distance-optimalDistance)/optimalDistance)*100.0)

			// Check if this connection is within the optimal range
			if math.Abs(distance-optimalDistance) <= 0.2*optimalDistance {
				connections = append(connections, Connection{
					ICAO:         fbo.ICAO,
					Distance:     distance,
					Contribution: contribution,
				})
			}
		}

		// Sort connections by contribution (higher is better)
		for i := 0; i < len(connections); i++ {
			for j := i + 1; j < len(connections); j++ {
				if connections[i].Contribution < connections[j].Contribution {
					connections[i], connections[j] = connections[j], connections[i]
				}
			}
		}

		// Limit to top 5 connections
		eligibleConnections := len(connections)
		displayLimit := 5
		if len(connections) > displayLimit {
			connections = connections[:displayLimit]
		}

		// Color code the score based on its value
		var coloredScore string
		intScore := int(score)
		if intScore >= 80 {
			coloredScore = green(fmt.Sprintf("%d", intScore))
		} else if intScore >= 50 {
			coloredScore = yellow(fmt.Sprintf("%d", intScore))
		} else {
			coloredScore = red(fmt.Sprintf("%d", intScore))
		}

		// Format with consistent column alignment for easier scanning like a table
		scoreSection := fmt.Sprintf("Score: %s", coloredScore)
		connectionsSection := fmt.Sprintf("Connections: %d/%d", eligibleConnections, totalConnections)

		result += fmt.Sprintf("%-3d %-40s  %-15s  %-20s\n",
			i+1,
			bold(airport.Name)+" "+cyan("("+airport.ICAO+")"),
			bold(scoreSection),
			bold(connectionsSection))

		// Add details about eligible connections (limited to top 5)
		if len(connections) > 0 {
			var connectionDetails []string
			for _, conn := range connections {
				connectionDetails = append(connectionDetails, fmt.Sprintf("%s (%d nm)", conn.ICAO, int(math.Round(conn.Distance))))
			}

			result += fmt.Sprintf("   %s\n",
				strings.Join(connectionDetails, ", "))
		}
	}

	return result, nil
}
