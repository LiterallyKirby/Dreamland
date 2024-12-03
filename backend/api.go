package quickTools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Install function simulates package installation
func Install(packageName string) {
	fmt.Println("Installing", packageName)
}

// PackageInfo represents details of an AUR package
type PackageInfo struct {
	Name        string  `json:"Name"`
	Description string  `json:"Description"`
	Version     string  `json:"Version"`
	Popularity  float32 `json:"Popularity"`
	Author      string  `json:"Maintainer"`
	URL         string  `json:"URL"`
}

// ApiResponse represents the API response structure
type ApiResponse struct {
	Results []PackageInfo `json:"results"`
}

// AUR RPC URL for searching packages
const aurURL = "https://aur.archlinux.org/rpc/?v=5&type=search&arg="

// Search function queries the AUR for packages matching the term
func Search(term string) ([]PackageInfo, error) {
	resp, err := http.Get(aurURL + term)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AUR: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data from AUR (status: %d)", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body) // Replacing ioutil.ReadAll with io.ReadAll
	if err != nil {
		return nil, fmt.Errorf("failed to read AUR response: %w", err)
	}

	var apiResponse ApiResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse AUR response: %w", err)
	}

	return apiResponse.Results, nil
}
