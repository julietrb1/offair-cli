package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"offair-cli/models/onair"
	"os"

	"offair-cli/models"
)

const onAirBaseURL = "https://server1.onair.company/api"
const onAirAuthHeaderName = "oa-apikey"

// OnAirAPI encapsulates the OnAir API client
type OnAirAPI struct {
	apiKey string
	client *http.Client
}

// NewOnAirAPI creates a new OnAir API client
func NewOnAirAPI() (*OnAirAPI, error) {
	apiKey := os.Getenv("ONAIR_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ONAIR_API_KEY is not set in the environment")
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &OnAirAPI{
		apiKey: apiKey,
		client: client,
	}, nil
}

// OAResponse is a reusable struct for handling OnAir responses.
type OAResponse[T any] struct {
	Content T `json:"Content"`
}

// GetAirport fetches an airport by its ICAO.
func (api *OnAirAPI) GetAirport(icao string) (*onair.Airport, error) {
	url := fmt.Sprintf("%s/v1/airports/%s", onAirBaseURL, icao)

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set(onAirAuthHeaderName, api.apiKey)

	// Perform HTTP request
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Validate response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received unexpected status code: %d", resp.StatusCode)
	}

	// Parse the API response
	var apiResp OAResponse[onair.Airport]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Validate the API response
	if apiResp.Content.ICAO == "" {
		return nil, fmt.Errorf("airport with ICAO %s not found in the API", icao)
	}

	if apiResp.Content.Name == "" {
		return nil, fmt.Errorf("airport with ICAO %s has no name in the API", icao)
	}

	// Note: We don't check for empty country code here anymore
	// If the country code is empty, we'll prompt the user to enter it

	return &apiResp.Content, nil
}

// GetAircraftType fetches an aircraft type by its ID.
func (api *OnAirAPI) GetAircraftType(aircraftTypeID string) (*models.AircraftType, error) {
	url := fmt.Sprintf("%s/v1/aircrafttypes/%s", onAirBaseURL, aircraftTypeID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(onAirAuthHeaderName, api.apiKey)

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResp OAResponse[models.AircraftType]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &apiResp.Content, nil
}

// GetAircraftAtAirport fetches all aircraft at a specific airport by its ICAO.
func (api *OnAirAPI) GetAircraftAtAirport(icao string) (*models.AircraftAtAirportResponse, error) {
	url := fmt.Sprintf("%s/v1/airports/%s/aircraft", onAirBaseURL, icao)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(onAirAuthHeaderName, api.apiKey)

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var aircraftResponse models.AircraftAtAirportResponse
	if err := json.NewDecoder(resp.Body).Decode(&aircraftResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &aircraftResponse, nil
}
