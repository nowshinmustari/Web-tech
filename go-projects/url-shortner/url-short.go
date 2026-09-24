
package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type URLRequest struct {
	URL string `json:"url"`
}

type URLResponse struct {
	ShortURL string `json:"short_url"`
}

var urls = make(map[string]string)
var mu sync.RWMutex

const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateShortCode() (string, error) {
    code := make([]byte, 6)

    for i := 0; i < 6; i++ {
        var randomByte [1]byte

        _, err := rand.Read(randomByte[:])
        if err != nil {
            return "", err
        }

        code[i] = characters[int(randomByte[0])%len(characters)]
    }

    shortCode := string(code)
    return shortCode, nil
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	var request URLRequest

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.URL == "" {
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(request.URL, "http://") &&
		!strings.HasPrefix(request.URL, "https://") {

		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	mu.Lock()

	shortCode, err := generateUniqueCode()

	if err != nil {
		mu.Unlock()
		http.Error(w, "Could not generate short code", http.StatusInternalServerError)
		return
	}

	urls[shortCode] = request.URL

	mu.Unlock()

	shortURL := "http://localhost:8080/" + shortCode

	response := URLResponse{
		ShortURL: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

func generateUniqueCode() (string, error) {
	for {
		code, err := generateShortCode()
		if err != nil {
			return "", err
		}

		if _, exists := urls[code]; !exists {
			return code, nil
		}
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {

	shortCode := strings.TrimPrefix(r.URL.Path, "/")

	mu.RLock()
	originalURL, exists := urls[shortCode]
	mu.RUnlock()

	if !exists {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func main() {

	http.HandleFunc("/shorten", shortenHandler)

	http.HandleFunc("/", redirectHandler)

	fmt.Println("Server running on http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Server error:", err)
	}
}