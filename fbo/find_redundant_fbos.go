package fbo

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"math"
	"offair-cli/models"
	"sort"
)

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
