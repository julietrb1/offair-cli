package menu

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"github.com/julietrb1/offair-cli/fbo"
	"os"
	"strconv"
)

// FindOptimalFBOLocations finds optimal locations for FBOs
func FindOptimalFBOLocations(db *sqlx.DB) {
	// Define color functions
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	// Get values from environment variables or use hardcoded defaults
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

	// Get FBO_REQ_LIGHTS environment variable (default to "true")
	requireLightsStr := os.Getenv("FBO_REQ_LIGHTS")
	if requireLightsStr == "" {
		requireLightsStr = "true" // Default value when environment variable is not set
	}
	requireLights := requireLightsStr == "true"

	// Get FBO_PREFERRED_SIZE environment variable (no default)
	preferredSizeStr := os.Getenv("FBO_PREFERRED_SIZE")
	var preferredSize *int
	if preferredSizeStr != "" {
		size, err := strconv.Atoi(preferredSizeStr)
		if err == nil && size >= 0 && size <= 5 {
			preferredSize = &size
		}
	}

	result, err := fbo.FindOptimalFBOLocations(db, optimalDistance, maxDistance, requireLights, preferredSize)
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Error:"), err)
		return
	}

	fmt.Println(bold(cyan("Calculating optimal FBO locations...")))
	fmt.Println(result)
}
