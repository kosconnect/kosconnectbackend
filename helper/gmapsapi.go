package helper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/organisasi/kosconnectbackend/models"
	"github.com/organisasi/kosconnectbackend/config"
)

// GeocodeResponse defines the structure of the response from HERE API
type GeocodeResponse struct {
	Items []struct {
		Position struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"position"`
	} `json:"items"`
}

// GetCoordinates retrieves the latitude and longitude for a given address
func GetCoordinates(address string) (float64, float64, error) {
	// Gunakan GetHereAPIURL untuk mendapatkan URL lengkap
	url := config.GetHereAPIURL(address)

	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch coordinates: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read response body: %v", err)
	}

	var result GeocodeResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse geocode response: %v", err)
	}

	if len(result.Items) == 0 {
		return 0, 0, fmt.Errorf("no coordinates found for address: %s", address)
	}

	return result.Items[0].Position.Lat, result.Items[0].Position.Lng, nil
}


// GetClosestPlaces retrieves nearby places from HERE Maps API
func GetClosestPlaces(longitude, latitude float64, apiKey, query string, maxDistance float64) ([]models.ClosestPlace, error) {
	// URL untuk HERE Places API
	url := fmt.Sprintf(
		"https://discover.search.hereapi.com/v1/discover?at=%f,%f&q=%s&apiKey=%s",
		latitude, longitude, query, apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch places: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse JSON response
	var result struct {
		Items []struct {
			Title    string  `json:"title"`
			Distance float64 `json:"distance"`
		} `json:"items"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %v", err)
	}

	// Filter places berdasarkan jarak
	places := []models.ClosestPlace{}
	for _, item := range result.Items {
		if item.Distance <= maxDistance*1000 { // Convert km to meters
			places = append(places, models.ClosestPlace{
				Name:     item.Title,
				Distance: item.Distance / 1000, // Convert meters to km
				Unit:     "km",
			})
		}
	}

	return places, nil
}
