package menu

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"github.com/julietrb1/offair-cli/fbo"
	"os"
	"strconv"
)

// FindRedundantFBOs finds FBOs that don't contribute significantly to the network
func FindRedundantFBOs(db *sqlx.DB) {
	optimalDistanceStr := os.Getenv("FBO_NM_OPTIMAL")
	if optimalDistanceStr == "" {
		optimalDistanceStr = "800" // Default value when environment variable is not set
	}
	optimalDistance, _ := strconv.ParseFloat(optimalDistanceStr, 64)

	maxDistanceStr := os.Getenv("FBO_NM_MAX")
	if maxDistanceStr == "" {
		maxDistanceStr = "1200" // Default value when environment variable is not set
	}
	maxDistance, _ := strconv.ParseFloat(maxDistanceStr, 64)

	requireLightsStr := os.Getenv("FBO_REQ_LIGHTS")
	if requireLightsStr == "" {
		requireLightsStr = "true" // Default value when environment variable is not set
	}
	requireLights := requireLightsStr == "true"

	preferredSizeStr := os.Getenv("FBO_PREFERRED_SIZE")
	var preferredSize *int
	if preferredSizeStr != "" {
		size, err := strconv.Atoi(preferredSizeStr)
		if err == nil && size >= 0 && size <= 5 {
			preferredSize = &size
		}
	}

	// Get FBO_REDUNDANCY_THRESHOLD environment variable (default to "100.0")
	redundancyThresholdStr := os.Getenv("FBO_REDUNDANCY_THRESHOLD")
	if redundancyThresholdStr == "" {
		redundancyThresholdStr = "100.0" // Default value when environment variable is not set
	}
	redundancyThreshold, _ := strconv.ParseFloat(redundancyThresholdStr, 64)

	result, err := fbo.FindRedundantFBOs(db, optimalDistance, maxDistance, requireLights, preferredSize, redundancyThreshold)
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Error:"), err)
		return
	}

	fmt.Println(result)
	fmt.Println()
}
