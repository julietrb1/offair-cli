package fbo

import (
	"fmt"
	"math"
	"sort"
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
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// First check total number of FBOs
	var totalFBOs []models.Airport
	err := db.Select(&totalFBOs, "SELECT * FROM airports WHERE has_fbo = TRUE")
	if err != nil {
		return "", fmt.Errorf("error fetching airports with FBOs: %w", err)
	}

	// Get FBOs with coordinates
	var fbos []models.FBO
	err = db.Select(&fbos, `
		SELECT f.id, f.airport_id, f.icao, f.name, a.latitude, a.longitude
		FROM fbos f
		JOIN airports a ON f.airport_id = a.id
	`)
	if err != nil {
		return "", fmt.Errorf("error fetching FBOs: %w", err)
	}

	if len(totalFBOs) < 2 {
		return bold(yellow("There are fewer than 2 FBOs in the network. No distance analysis possible.")), nil
	}

	if len(fbos) < 2 {
		return bold(yellow(fmt.Sprintf(
			"Found %d FBOs in total, but fewer than 2 have valid coordinate information. "+
				"At least 2 FBOs with coordinates are needed to calculate distances.",
			len(totalFBOs)))), nil
	}

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

// FindRedundantFBOs identifies FBOs that don't contribute significantly to the overall network
// Uses a redundancy threshold (default 100.0) and a small co-location distance (10nm)
// to identify redundant FBOs. The algorithm uses a stable scoring system that produces
// consistent results across different threshold values, making it more predictable and
// less sensitive to small changes in the threshold.
func FindRedundantFBOs(db *sqlx.DB, optimalDistance, maxDistance float64, requireLights bool, preferredSize *int, redundancyThreshold float64) (string, error) {
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// First check total number of FBOs without filtering for lat/long
	var totalFBOs []models.Airport
	err := db.Select(&totalFBOs, "SELECT * FROM airports WHERE has_fbo = TRUE")
	if err != nil {
		return "", fmt.Errorf("error fetching existing FBOs: %w", err)
	}

	if len(totalFBOs) < 2 {
		return bold(yellow("There are fewer than 2 FBOs in the network. No redundancy analysis possible.")), nil
	}

	// Get existing FBOs with valid coordinates
	var existingFBOs []models.Airport
	err = db.Select(&existingFBOs, "SELECT * FROM airports WHERE has_fbo = TRUE AND latitude IS NOT NULL AND longitude IS NOT NULL")
	if err != nil {
		return "", fmt.Errorf("error fetching existing FBOs: %w", err)
	}

	if len(existingFBOs) < 2 {
		return bold(yellow(fmt.Sprintf(
			"Found %d FBOs in total, but only %d have valid latitude/longitude information. "+
				"At least 2 FBOs with coordinates are needed for redundancy analysis.",
			len(totalFBOs), len(existingFBOs)))), nil
	}

	// Calculate initial network metrics
	initialMetrics, err := calculateNetworkMetrics(existingFBOs, optimalDistance)
	if err != nil {
		return bold(yellow(fmt.Sprintf(
			"Error calculating network metrics: %v", err))), nil
	}

	// Structure to hold FBO scores
	type FBOScore struct {
		FBO   models.Airport
		Score float64
	}

	// Recursive function to find redundant FBOs
	var findRedundantFBOsRecursive func(fboList []models.Airport) ([]FBOScore, error)
	findRedundantFBOsRecursive = func(fboList []models.Airport) ([]FBOScore, error) {
		if len(fboList) < 2 {
			return nil, nil
		}

		var fboScores []FBOScore

		// For each FBO, calculate the impact of removing it
		for i, fbo := range fboList {
			// Create a new list without this FBO
			remainingFBOs := make([]models.Airport, 0, len(fboList)-1)
			remainingFBOs = append(remainingFBOs, fboList[:i]...)
			remainingFBOs = append(remainingFBOs, fboList[i+1:]...)

			// Calculate network metrics without this FBO
			metrics, err := calculateNetworkMetrics(remainingFBOs, optimalDistance)
			if err != nil {
				// Skip this FBO if we can't calculate metrics without it
				continue
			}

			// Calculate redundancy score (higher means more redundant)
			// A positive score means the network improves when this FBO is removed
			// A negative score means the network gets worse when this FBO is removed
			score := calculateRedundancyScore(initialMetrics, metrics)

			// Apply size preference if specified
			if preferredSize != nil && fbo.Size != nil {
				size := *fbo.Size
				preferredSizeVal := *preferredSize

				if size != preferredSizeVal {
					// Increase redundancy score for FBOs that don't match preferred size
					if size == preferredSizeVal+1 || size == preferredSizeVal-1 {
						// Smaller penalty for sizes close to preferred
						score += 5.0
					} else {
						// Larger penalty for sizes far from preferred
						score += 10.0
					}
				}
			}

			// Apply negative weight for airports with lights if required
			if requireLights && fbo.HasLights {
				// Decrease redundancy score for FBOs with lights when lights are required
				score -= 10.0
			}

			fboScores = append(fboScores, FBOScore{
				FBO:   fbo,
				Score: score,
			})
		}

		// Sort FBOs by redundancy score (higher is more redundant)
		sort.Slice(fboScores, func(i, j int) bool {
			return fboScores[i].Score > fboScores[j].Score
		})

		return fboScores, nil
	}

	// Find redundant FBOs recursively until the network stabilizes
	var allRedundantFBOs []FBOScore
	currentFBOs := make([]models.Airport, len(existingFBOs))
	copy(currentFBOs, existingFBOs)

	// Keep track of removed FBOs to avoid co-located redundancy
	removedFBOIDs := make(map[string]bool)

	// Recursive removal until no more beneficial removals are found
	for {
		redundantFBOs, err := findRedundantFBOsRecursive(currentFBOs)
		if err != nil {
			return "", err
		}

		// If no redundant FBOs or no scores above threshold, we're done
		if len(redundantFBOs) == 0 || redundantFBOs[0].Score <= redundancyThreshold {
			break
		}

		// Add the most redundant FBO to our list if it has a score above threshold
		if redundantFBOs[0].Score > redundancyThreshold {
			// Check if this FBO is co-located with any already removed FBO
			isColocated := false
			for _, existingRedundant := range allRedundantFBOs {
				if existingRedundant.FBO.Latitude != nil && existingRedundant.FBO.Longitude != nil &&
					redundantFBOs[0].FBO.Latitude != nil && redundantFBOs[0].FBO.Longitude != nil {

					distance := CalculateDistance(
						*existingRedundant.FBO.Latitude, *existingRedundant.FBO.Longitude,
						*redundantFBOs[0].FBO.Latitude, *redundantFBOs[0].FBO.Longitude)

					// If FBOs are within 10nm, consider them co-located
					if distance < 10 {
						isColocated = true
						break
					}
				}
			}

			if !isColocated && !removedFBOIDs[redundantFBOs[0].FBO.ID] {
				allRedundantFBOs = append(allRedundantFBOs, redundantFBOs[0])
				removedFBOIDs[redundantFBOs[0].FBO.ID] = true

				// Remove this FBO from the current list for next iteration
				for i, fbo := range currentFBOs {
					if fbo.ID == redundantFBOs[0].FBO.ID {
						currentFBOs = append(currentFBOs[:i], currentFBOs[i+1:]...)
						break
					}
				}
			} else {
				// If co-located or already removed, skip to next best candidate
				if len(redundantFBOs) > 1 {
					for i := 1; i < len(redundantFBOs); i++ {
						if redundantFBOs[i].Score <= redundancyThreshold {
							break // No more scores above threshold
						}

						isNextColocated := false
						for _, existingRedundant := range allRedundantFBOs {
							if existingRedundant.FBO.Latitude != nil && existingRedundant.FBO.Longitude != nil &&
								redundantFBOs[i].FBO.Latitude != nil && redundantFBOs[i].FBO.Longitude != nil {

								distance := CalculateDistance(
									*existingRedundant.FBO.Latitude, *existingRedundant.FBO.Longitude,
									*redundantFBOs[i].FBO.Latitude, *redundantFBOs[i].FBO.Longitude)

								if distance < 10 {
									isNextColocated = true
									break
								}
							}
						}

						if !isNextColocated && !removedFBOIDs[redundantFBOs[i].FBO.ID] {
							allRedundantFBOs = append(allRedundantFBOs, redundantFBOs[i])
							removedFBOIDs[redundantFBOs[i].FBO.ID] = true

							// Remove this FBO from the current list for next iteration
							for j, fbo := range currentFBOs {
								if fbo.ID == redundantFBOs[i].FBO.ID {
									currentFBOs = append(currentFBOs[:j], currentFBOs[j+1:]...)
									break
								}
							}
							break
						}
					}
				}
			}
		}

		// If we couldn't find any non-colocated FBO to remove or if we're down to minimum FBOs, stop
		if len(currentFBOs) < 2 || len(allRedundantFBOs) == 0 || len(allRedundantFBOs) == len(existingFBOs)-1 {
			break
		}
	}

	// Build result string in scenario format
	result := fmt.Sprintf("%s\n\n", bold(cyan("FBO Redundancy Analysis:")))

	// Add configuration information
	result += fmt.Sprintf("%s %s %.2f nm, %s %.2f nm\n",
		bold("Using:"),
		bold("optimal distance:"), optimalDistance,
		bold("maximum distance:"), maxDistance)

	// Add information about lights requirement
	if requireLights {
		result += fmt.Sprintf("%s %s\n",
			bold("Requiring airports with lights:"),
			green("Yes"))
	} else {
		result += fmt.Sprintf("%s %s\n",
			bold("Requiring airports with lights:"),
			yellow("No"))
	}

	// Add information about preferred size if specified
	if preferredSize != nil {
		result += fmt.Sprintf("%s %d\n",
			bold("Preferred airport size:"),
			*preferredSize)
	}

	// Add information about redundancy threshold
	result += fmt.Sprintf("%s %.1f %s\n",
		bold("Redundancy threshold:"),
		redundancyThreshold,
		yellow("(scores range from 0-100, higher threshold = less aggressive)"))

	result += fmt.Sprintf("%s %d existing FBOs in the network.\n\n",
		bold("Found:"), len(existingFBOs))

	// If no redundant FBOs were found
	if len(allRedundantFBOs) == 0 {
		result += bold(green("Scenario Assessment: ")) + fmt.Sprintf(
			"Based on the analysis with a redundancy threshold of %.1f, no FBOs are considered redundant in the current network. "+
				"The existing FBO distribution provides optimal coverage given the specified criteria.\n\n"+
				"No changes are recommended at this time. If you wish to identify more FBOs for potential removal, "+
				"you can lower the redundancy threshold by setting the FBO_REDUNDANCY_THRESHOLD environment variable.\n\n"+
				"The redundancy score ranges from 0 to 100, with higher scores indicating FBOs that contribute less to the network. "+
				"A threshold of 100 is very strict (no FBOs will be considered redundant), while a threshold of 50 is moderate, "+
				"and a threshold of 0 would consider all FBOs for potential removal (not recommended).",
			redundancyThreshold)
		return result, nil
	}

	// Calculate metrics for the optimized network
	optimizedFBOs := make([]models.Airport, 0, len(existingFBOs)-len(allRedundantFBOs))
	for _, fbo := range existingFBOs {
		if !removedFBOIDs[fbo.ID] {
			optimizedFBOs = append(optimizedFBOs, fbo)
		}
	}

	optimizedMetrics, err := calculateNetworkMetrics(optimizedFBOs, optimalDistance)
	if err != nil {
		return "", err
	}

	// Add scenario description
	result += bold(green("Scenario Assessment: ")) + fmt.Sprintf(
		"The analysis identified %d FBOs that could be considered redundant without significantly impacting network coverage. "+
			"With the current redundancy threshold of %.1f, only FBOs with scores above this value are considered for removal, "+
			"ensuring that only the most redundant FBOs are identified while maintaining adequate network coverage.\n\n"+
			"The redundancy score ranges from 0 to 100, with higher scores indicating FBOs that contribute less to the network. "+
			"The scores are calculated using a stable algorithm that considers how each FBO affects the overall network metrics "+
			"when removed. This approach ensures consistent results across different threshold values.\n\n",
		len(allRedundantFBOs), redundancyThreshold)

	// Add before/after metrics comparison
	result += bold("Network Metrics Comparison:\n")

	// Calculate FBO count change
	fboChangeSymbol := "-"
	if len(optimizedFBOs) >= len(existingFBOs) {
		fboChangeSymbol = "+"
	}
	fboChangePercent := math.Abs(float64(len(optimizedFBOs)-len(existingFBOs)) / float64(len(existingFBOs)) * 100)

	result += fmt.Sprintf("  • %s: %d → %d (%s%.0f%%)\n",
		bold("Total FBOs"),
		len(existingFBOs),
		len(optimizedFBOs),
		fboChangeSymbol,
		fboChangePercent)

	// Calculate average distance change
	distChangeSymbol := "-"
	if optimizedMetrics.averageDistance > initialMetrics.averageDistance {
		distChangeSymbol = "+"
	}
	distChangePercent := math.Abs((optimizedMetrics.averageDistance - initialMetrics.averageDistance) / initialMetrics.averageDistance * 100)

	result += fmt.Sprintf("  • %s: %.2f nm → %.2f nm (%s%.2f%%)\n",
		bold("Average distance between FBOs"),
		initialMetrics.averageDistance,
		optimizedMetrics.averageDistance,
		distChangeSymbol,
		distChangePercent)

	// Calculate efficiency score change
	effChangeSymbol := "-"
	if optimizedMetrics.efficiencyScore > initialMetrics.efficiencyScore {
		effChangeSymbol = "+"
	}
	effChangePercent := math.Abs((optimizedMetrics.efficiencyScore - initialMetrics.efficiencyScore) / initialMetrics.efficiencyScore * 100)

	result += fmt.Sprintf("  • %s: %.2f → %.2f (%s%.2f%%)\n\n",
		bold("Network efficiency score"),
		initialMetrics.efficiencyScore,
		optimizedMetrics.efficiencyScore,
		effChangeSymbol,
		effChangePercent)

	// List redundant FBOs
	result += bold(yellow("Recommended FBOs for removal:")) + "\n"
	for i, fboScore := range allRedundantFBOs {
		fbo := fboScore.FBO

		// Color code the score based on its value
		var coloredScore string
		if fboScore.Score >= 20 {
			coloredScore = red(fmt.Sprintf("%.1f", fboScore.Score))
		} else if fboScore.Score >= 10 {
			coloredScore = yellow(fmt.Sprintf("%.1f", fboScore.Score))
		} else {
			coloredScore = green(fmt.Sprintf("%.1f", fboScore.Score))
		}

		result += fmt.Sprintf("%d. %s %s - Redundancy Score: %s\n",
			i+1,
			bold(fbo.Name),
			cyan("("+fbo.ICAO+")"),
			coloredScore)

		// Find nearest remaining FBOs
		var nearestFBOs []struct {
			ICAO     string
			Distance float64
		}

		for _, remainingFBO := range optimizedFBOs {
			if fbo.Latitude != nil && fbo.Longitude != nil &&
				remainingFBO.Latitude != nil && remainingFBO.Longitude != nil {

				distance := CalculateDistance(
					*fbo.Latitude, *fbo.Longitude,
					*remainingFBO.Latitude, *remainingFBO.Longitude)

				nearestFBOs = append(nearestFBOs, struct {
					ICAO     string
					Distance float64
				}{
					ICAO:     remainingFBO.ICAO,
					Distance: distance,
				})
			}
		}

		// Sort by distance
		sort.Slice(nearestFBOs, func(i, j int) bool {
			return nearestFBOs[i].Distance < nearestFBOs[j].Distance
		})

		// Show up to 3 nearest FBOs
		if len(nearestFBOs) > 0 {
			limit := 3
			if len(nearestFBOs) < limit {
				limit = len(nearestFBOs)
			}

			result += fmt.Sprintf("   Nearest alternative FBOs: ")
			for j := 0; j < limit; j++ {
				if j > 0 {
					result += ", "
				}
				result += fmt.Sprintf("%s (%.0f nm)", nearestFBOs[j].ICAO, nearestFBOs[j].Distance)
			}
			result += "\n"
		}
	}

	return result, nil
}

// NetworkMetrics holds metrics about the FBO network
type NetworkMetrics struct {
	averageDistance    float64
	efficiencyScore    float64
	optimalConnections int
	totalConnections   int
}

// calculateNetworkMetrics calculates various metrics for the FBO network
func calculateNetworkMetrics(fboList []models.Airport, optimalDistance float64) (NetworkMetrics, error) {
	if len(fboList) < 2 {
		return NetworkMetrics{}, fmt.Errorf("need at least 2 FBOs to calculate network metrics")
	}

	var metrics NetworkMetrics
	var totalDistance float64
	var connections int
	var optimalConnections int
	var fbosWithValidCoords int

	// Count FBOs with valid coordinates
	for _, fbo := range fboList {
		if fbo.Latitude != nil && fbo.Longitude != nil {
			fbosWithValidCoords++
		}
	}

	if fbosWithValidCoords < 2 {
		return NetworkMetrics{}, fmt.Errorf("need at least 2 FBOs with valid coordinates to calculate network metrics, found %d", fbosWithValidCoords)
	}

	// Calculate distances between all FBOs
	for i := 0; i < len(fboList); i++ {
		for j := i + 1; j < len(fboList); j++ {
			if fboList[i].Latitude == nil || fboList[i].Longitude == nil ||
				fboList[j].Latitude == nil || fboList[j].Longitude == nil {
				continue
			}

			distance := CalculateDistance(
				*fboList[i].Latitude, *fboList[i].Longitude,
				*fboList[j].Latitude, *fboList[j].Longitude)

			totalDistance += distance
			connections++

			// Check if this connection is within 20% of the optimal distance
			if math.Abs(distance-optimalDistance) <= 0.2*optimalDistance {
				optimalConnections++
			}
		}
	}

	if connections == 0 {
		return NetworkMetrics{}, fmt.Errorf("found %d FBOs with valid coordinates but no valid connections between them", fbosWithValidCoords)
	}

	// Calculate average distance
	metrics.averageDistance = totalDistance / float64(connections)

	// Calculate efficiency score (higher is better)
	// Based on how many connections are optimal and how close the average is to optimal
	optimalRatio := float64(optimalConnections) / float64(connections)
	distanceScore := 100.0 - math.Min(100.0, (math.Abs(metrics.averageDistance-optimalDistance)/optimalDistance)*100.0)

	metrics.efficiencyScore = (optimalRatio * 50.0) + (distanceScore * 0.5)
	metrics.optimalConnections = optimalConnections
	metrics.totalConnections = connections

	return metrics, nil
}

// calculateRedundancyScore calculates how redundant an FBO is
// Higher score means more redundant (better candidate for removal)
func calculateRedundancyScore(originalMetrics, newMetrics NetworkMetrics) float64 {
	// Calculate percentage changes
	avgDistanceChange := (newMetrics.averageDistance - originalMetrics.averageDistance) / originalMetrics.averageDistance
	efficiencyChange := (newMetrics.efficiencyScore - originalMetrics.efficiencyScore) / originalMetrics.efficiencyScore

	// Calculate optimal connection ratio change
	originalRatio := float64(originalMetrics.optimalConnections) / float64(originalMetrics.totalConnections)
	newRatio := float64(newMetrics.optimalConnections) / float64(newMetrics.totalConnections)
	ratioChange := newRatio - originalRatio

	// Combine factors into a score
	// Positive score means the FBO is redundant (network improves when removed)
	// Negative score means the FBO is important (network gets worse when removed)

	// Apply logarithmic scaling to make the algorithm more stable
	// This will spread out the scores more evenly and reduce sensitivity to small changes
	efficiencyComponent := math.Log1p(math.Abs(efficiencyChange)) * 20.0
	if efficiencyChange < 0 {
		efficiencyComponent = -efficiencyComponent
	}

	ratioComponent := math.Log1p(math.Abs(ratioChange)) * 15.0
	if ratioChange < 0 {
		ratioComponent = -ratioComponent
	}

	distanceComponent := math.Log1p(math.Abs(avgDistanceChange)) * 10.0
	if avgDistanceChange > 0 {
		distanceComponent = -distanceComponent
	}

	// Combine the components with a sigmoid function to smooth the transition
	rawScore := efficiencyComponent + ratioComponent + distanceComponent

	// Apply sigmoid scaling to create a more gradual transition around the threshold
	// This transforms the score to a range of approximately 0-100
	score := 100.0 / (1.0 + math.Exp(-rawScore/10.0))

	return score
}
