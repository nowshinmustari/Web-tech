package main

import (
	"fmt"
	"net/http"
)

// In-memory URL storage
var urlStore = make(map[string]string)
var counter = 0

func shortenHandler(w http.ResponseWriter, r *http.Request) {

	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	longURL := r.FormValue("url")

	if longURL == "" {
		http.Error(w, "url is missing", http.StatusBadRequest)
		return
	}

	counter++
	shortCode := fmt.Sprintf("%d", counter)

	urlStore[shortCode] = longURL

	fmt.Fprintf(w, "Short URL created: http://localhost:8080/%s\n", shortCode)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {

	shortCode := r.URL.Path[1:]
	longURL, found := urlStore[shortCode]

	if !found {
		http.Error(w, "This short URL was not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}

func main() {

	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/", redirectHandler)

	fmt.Println("Server running at: http://localhost:8080")

	http.ListenAndServe(":8080", nil)
}