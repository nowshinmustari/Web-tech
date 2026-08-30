package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type User struct {
	Username string
	Password string
}

// For learning purpose only.
// In a real application, users should come from a database
// and passwords should be stored as bcrypt hashes.
var users = map[string]User{
	"admin": {
		Username: "admin",
		Password: "123456",
	},
}

// In-memory token storage.
// token -> username
var tokens = map[string]string{
	"abc123": "admin",
}

func healthHandler(w http.ResponseWriter, r *http.Request) {

	response := map[string]interface{}{
		"msg":    "ok",
		"status": 200,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error encoding response:", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	// Only POST request is allowed
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Request body structure
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Convert JSON request body into Go struct
	if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Find user
	user, exists := users[loginData.Username]

	if !exists || user.Password != loginData.Password {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// For learning purpose, token is fixed.
	// In a real application, generate a secure random token/JWT.
	token := "abc123"

	// Save token
	tokens[token] = user.Username

	response := map[string]interface{}{
		"msg":   "Login successful",
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")

		// Example:
		// Authorization: Bearer abc123

		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check "Bearer " prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Check token
		username, exists := tokens[token]

		if !exists {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		log.Println("Authenticated user:", username)

		// Authentication successful
		next(w, r)
	}
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {

	response := map[string]interface{}{
		"msg":    "You are authenticated",
		"status": 200,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

func main() {

	// Public endpoint
	http.HandleFunc("/api/v1/health", healthHandler)

	// Login endpoint
	http.HandleFunc("/api/v1/login", loginHandler)

	// Protected endpoint
	http.HandleFunc(
		"/api/v1/profile",
		authMiddleware(protectedHandler),
	)

	log.Println("Server running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}