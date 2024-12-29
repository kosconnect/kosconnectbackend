package config

import (
	"fmt"
	"os"
)

func GetGitHubToken() string {
	return os.Getenv("GH_ACCESS_TOKEN")
}

// ini hanya untuk nyari longitude dan latitude berdasarkan alamat saja
func GetHereAPIURL(address string) string {
	apiKey := os.Getenv("HERE_API_KEY")
	if apiKey == "" {
		panic("HERE_API_KEY is not set in environment variables")
	}
	return fmt.Sprintf("https://geocode.search.hereapi.com/v1/geocode?q=%s&apiKey=%s", address, apiKey)
}

//ini untuk apapun
func GetHereAPIKey() string {
	apiKey := os.Getenv("HERE_API_KEY")
	if apiKey == "" {
		panic("HERE_API_KEY is not set in environment variables")
	}
	return apiKey
}
