package fbo

import (
	"fmt"
	"math"
	"strings"

	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"

	"offair-cli/models"
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

// CalculateDistance calculates the distance between two points using the Haversine formula
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusNM = 3440.0 // Earth radius in nm

	// Convert latitude and longitude from degrees to radians
	lat1Rad := lat1 * (math.Pi / 180.0)
	lon1Rad := lon1 * (math.Pi / 180.0)
	lat2Rad := lat2 * (math.Pi / 180.0)
	lon2Rad := lon2 * (math.Pi / 180.0)

	// Haversine formula
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadiusNM * c

	return distance
}

// ListDistancesBetweenFBOs lists the distances between all FBOs in a more organized and insightful way
func ListDistancesBetweenFBOs(db *sqlx.DB) (string, error) {
	var fbos []models.FBO
	err := db.Select(&fbos, `
		SELECT f.id, f.airport_id, f.icao, f.name, a.latitude, a.longitude
		FROM fbos f
		JOIN airports a ON f.airport_id = a.id
	`)
	if err != nil {
		return "", fmt.Errorf("error fetching FBOs: %w", err)
	}

	if len(fbos) < 2 {
		return "", fmt.Errorf("need at least 2 FBOs to calculate distances")
	}

	// Define color functions
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Calculate all distances
	type DistancePair struct {
		FBO1     models.FBO
		FBO2     models.FBO
		Distance float64
	}

	var distances []DistancePair
	var totalDistance float64
	var minDistance = math.MaxFloat64
	var maxDistance float64
	var minPair, maxPair DistancePair

	for i := 0; i < len(fbos); i++ {
		for j := i + 1; j < len(fbos); j++ {
			distance := CalculateDistance(fbos[i].Latitude, fbos[i].Longitude, fbos[j].Latitude, fbos[j].Longitude)
			pair := DistancePair{
				FBO1:     fbos[i],
				FBO2:     fbos[j],
				Distance: distance,
			}
			distances = append(distances, pair)
			totalDistance += distance

			// Track min and max distances
			if distance < minDistance {
				minDistance = distance
				minPair = pair
			}
			if distance > maxDistance {
				maxDistance = pair.Distance
				maxPair = pair
			}
		}
	}

	// Sort distances from shortest to longest
	for i := 0; i < len(distances); i++ {
		for j := i + 1; j < len(distances); j++ {
			if distances[i].Distance > distances[j].Distance {
				distances[i], distances[j] = distances[j], distances[i]
			}
		}
	}

	// Calculate average distance
	avgDistance := totalDistance / float64(len(distances))

	// Build result string
	result := fmt.Sprintf("%s\n\n", bold(cyan("FBO Network Analysis:")))

	// Add summary statistics
	result += fmt.Sprintf("%s\n", bold("Summary Statistics:"))
	result += fmt.Sprintf("  • %s: %d\n", bold("Total FBOs"), len(fbos))
	result += fmt.Sprintf("  • %s: %d\n", bold("Total connections"), len(distances))
	result += fmt.Sprintf("  • %s: %.2f nm\n", bold("Average distance"), avgDistance)
	result += fmt.Sprintf("  • %s: %.2f nm (%s to %s)\n",
		bold("Shortest connection"),
		minDistance,
		bold(minPair.FBO1.ICAO),
		bold(minPair.FBO2.ICAO))
	result += fmt.Sprintf("  • %s: %.2f nm (%s to %s)\n\n",
		bold("Longest connection"),
		maxDistance,
		bold(maxPair.FBO1.ICAO),
		bold(maxPair.FBO2.ICAO))

	// Add closest connections section
	result += fmt.Sprintf("%s\n", bold(green("Closest Connections:")))
	limit := 5
	if len(distances) < limit {
		limit = len(distances)
	}
	for i := 0; i < limit; i++ {
		result += fmt.Sprintf("  %d. %s %s %s: %.2f nm\n",
			i+1,
			bold(distances[i].FBO1.ICAO),
			blue("to"),
			bold(distances[i].FBO2.ICAO),
			distances[i].Distance)
	}
	result += "\n"

	// Add furthest connections section
	result += fmt.Sprintf("%s\n", bold(red("Furthest Connections:")))
	start := len(distances) - limit
	if start < 0 {
		start = 0
	}
	for i := start; i < len(distances); i++ {
		result += fmt.Sprintf("  %d. %s %s %s: %.2f nm\n",
			i-start+1,
			bold(distances[i].FBO1.ICAO),
			blue("to"),
			bold(distances[i].FBO2.ICAO),
			distances[i].Distance)
	}
	result += "\n"

	// Group FBOs by proximity
	// We'll create clusters based on distance thresholds
	result += fmt.Sprintf("%s\n", bold(yellow("FBO Clusters:")))

	// Define distance threshold for clustering
	shortDistance := 300.0 // nm

	// Find clusters of closely located FBOs
	visited := make(map[string]bool)
	clusterCount := 0

	for i := 0; i < len(fbos); i++ {
		if visited[fbos[i].ICAO] {
			continue
		}

		// Start a new cluster
		var cluster []string
		cluster = append(cluster, fbos[i].ICAO)
		visited[fbos[i].ICAO] = true

		// Find all FBOs close to this one
		for j := 0; j < len(fbos); j++ {
			if i == j || visited[fbos[j].ICAO] {
				continue
			}

			distance := CalculateDistance(fbos[i].Latitude, fbos[i].Longitude, fbos[j].Latitude, fbos[j].Longitude)
			if distance <= shortDistance {
				cluster = append(cluster, fbos[j].ICAO)
				visited[fbos[j].ICAO] = true
			}
		}

		// Only show clusters with at least 2 FBOs
		if len(cluster) >= 2 {
			clusterCount++
			result += fmt.Sprintf("  %s %d: %s (within %.0f nm)\n",
				bold("Cluster"),
				clusterCount,
				strings.Join(cluster, ", "),
				shortDistance)
		}
	}

	// If no clusters were found
	if clusterCount == 0 {
		result += fmt.Sprintf("  %s\n", yellow("No clusters found within "+fmt.Sprintf("%.0f", shortDistance)+" nm"))
	}

	// Add a note about viewing all distances
	if len(distances) > 10 {
		result += fmt.Sprintf("\n%s %d %s\n",
			bold(yellow("Note:")),
			len(distances),
			yellow("total connections exist. Only the most significant are shown above."))
	}

	return result, nil
}

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
