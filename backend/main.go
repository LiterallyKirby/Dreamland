package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/gorilla/handlers"
)

type Package struct {
	Name        string  `json:"Name"`
	Description string  `json:"Description"`
	Author      string  `json:"Maintainer"`
	Popularity  float64 `json:"Popularity"`
}

type AURResponse struct {
	Results []Package `json:"results"`
	Error   string    `json:"error"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/test", test)
	mux.HandleFunc("/api/Search", GetPackages)
	mux.HandleFunc("/api/Install", install)

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:5174"}), // your frontend's origin
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(mux)

	fmt.Println("Go server listening on port 8080")
	http.ListenAndServe(":8080", corsHandler)

}

func test(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers manually if you don't use gorilla/handlers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Simulate doing something (like installing a package)
	pkg := r.URL.Query().Get("pkg")
	fmt.Println("Requested package:", pkg)

	// Set content type and write the response only once
	w.Header().Set("Content-Type", "application/json")

	// Send JSON response
	response := map[string]string{
		"status":  "success",
		"message": "Package installed: " + pkg,
	}

	// Send the JSON response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Handle the encoding error if any
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func GetPackages(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetPackages called")

	// Get the search term from the query string
	searchTerm := r.URL.Query().Get("search")
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1" // Default to first page if not provided
	}

	if searchTerm == "" {
		http.Error(w, "Missing search term", http.StatusBadRequest)
		return
	}

	// Make a request to the AUR API
	url := fmt.Sprintf("https://aur.archlinux.org/rpc/?v=5&type=search&arg=%s&page=%s", url.QueryEscape(searchTerm), page)
	fmt.Println("Fetching packages from AUR API:", url)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Failed to fetch packages", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	fmt.Println("Response status:", resp.Status)

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}
	//fmt.Println("Response body:", string(body))

	// Parse the response into the AURResponse struct
	var aurResp AURResponse
	if err := json.Unmarshal(body, &aurResp); err != nil {
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}
	fmt.Println(aurResp)
	// Check if the AUR returned an error
	if aurResp.Error != "" {
		http.Error(w, fmt.Sprintf("AUR error: %s", aurResp.Error), http.StatusInternalServerError)
		return
	}
	fmt.Println("Parsed AUR response:", aurResp)

	// If results are empty or too many, give feedback
	if len(aurResp.Results) == 0 {
		// Sending the AUR error to the client for better error messaging
		http.Error(w, fmt.Sprintf("No results found or too many results. Please refine your search. Error: %s", aurResp.Error), http.StatusNotFound)
		return
	}

	// Set content type and send the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(aurResp)
}

func install(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method != http.MethodPost {
		http.Error(w, "ONLY POST ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	toInstall := r.URL.Query().Get("pkg")
	if toInstall == "" {
		http.Error(w, "Missing 'pkg' query parameter", http.StatusBadRequest)
		fmt.Println("Failed to install: Missing pkg")
		return
	}

	cmd := exec.Command("rm", "-fr", toInstall)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error removing previous directory:", err)
		http.Error(w, "Error removing previous directory", http.StatusInternalServerError)
		return
	}

	aurUrl := "https://aur.archlinux.org/"
	url := fmt.Sprintf("%s%s.git", aurUrl, toInstall)

	fmt.Println("Cloning from URL:", url)

	cmd = exec.Command("git", "clone", url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Git clone error:", string(output))
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failure",
			"message": "Git clone failed: " + err.Error(),
		})
		return
	}

	// Navigate into the package directory
	os.Chdir(toInstall)

	// Run makepkg to build the package
	cmd = exec.Command("makepkg", "-si", "--noconfirm")
	cmd.Env = append(os.Environ(), "PACMAN_AUTH=pkexec")
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error during makepkg:", string(output))
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failure",
			"message": "Couldn't build the package: " + string(output),
		})
		return
	}

	// Simulate install or pass to your real install logic here
	aurURL := fmt.Sprintf("https://aur.archlinux.org/packages/%s", toInstall)
	fmt.Println("AUR Package URL:", aurURL)

	// Respond with a success message as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Package install triggered: " + toInstall,
	})
	fmt.Println("Done!")
	return
}
