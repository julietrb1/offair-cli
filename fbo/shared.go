package fbo

import (
	"fmt"
	"github.com/julietrb1/offair-cli/models"
	"math"
)

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
