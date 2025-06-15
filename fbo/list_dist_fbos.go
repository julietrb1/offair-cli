package fbo

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"math"
	"offair-cli/models"
	"strings"
)

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
