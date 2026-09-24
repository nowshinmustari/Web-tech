package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)


type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// =========================
// USER
// =========================

type User struct {
	Username string
	Password string
}

// Learning purpose only
var users = map[string]User{
	"admin": {
		Username: "admin",
		Password: "123456",
	},
}

// JWT secret key
var jwtSecret = []byte("my-super-secret-key")

// =========================
// LOGIN RESPONSE
// =========================

type LoginResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

// =========================
// HEALTH HANDLER
// =========================

func healthHandler(w http.ResponseWriter, r *http.Request) {

	response := map[string]interface{}{
		"msg":    "ok",
		"status": 200,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error encoding response:", err)
	}
}

// =========================
// GENERATE JWT
// =========================

func generateToken(username string) (string, error) {

	// Token expiration: 15 minutes
	expiresAt := time.Now().Add(15 * time.Minute)

	claims := Claims{
		Username: username,

		RegisteredClaims: jwt.RegisteredClaims{

			// Token expiration time
			ExpiresAt: jwt.NewNumericDate(expiresAt),

			// Token issued time
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create JWT
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	// Sign JWT
	tokenString, err := token.SignedString(jwtSecret)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// =========================
// VALIDATE JWT
// =========================

func validateToken(tokenString string) (*Claims, error) {

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,

		func(token *jwt.Token) (interface{}, error) {

			// Make sure token uses HS256
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}

			return jwtSecret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

// =========================
// LOGIN HANDLER
// =========================

func loginHandler(w http.ResponseWriter, r *http.Request) {
 
	// Only POST allowed
	if r.Method != http.MethodPost {

		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)

		return
	}

	// Request body
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Decode JSON
	err := json.NewDecoder(r.Body).Decode(&loginData)

	if err != nil {

		log.Println("JSON decode error:", err)

		http.Error(
			w,
			"Invalid JSON",
			http.StatusBadRequest,
		)

		return
	}

	// Find user
	user, exists := users[loginData.Username]

	// Check username and password
	if !exists || user.Password != loginData.Password {

		http.Error(
			w,
			"Invalid username or password",
			http.StatusUnauthorized,
		)

		return
	}

	// Generate JWT
	token, err := generateToken(user.Username)

	if err != nil {

		log.Println("JWT generation error:", err)

		http.Error(
			w,
			"Failed to generate token",
			http.StatusInternalServerError,
		)

		return
	}

	// Response
	response := LoginResponse{
		Status:  "OK",
		Message: "Login successful",
		Token:   token,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

// =========================
// AUTH MIDDLEWARE
// =========================

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")

		// Example:
		// Authorization: Bearer eyJhbGciOiJIUzI1Ni...

		if authHeader == "" {

			http.Error(
				w,
				"Authorization header required",
				http.StatusUnauthorized,
			)

			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {

			http.Error(
				w,
				"Invalid authorization format",
				http.StatusUnauthorized,
			)

			return
		}

		// Extract JWT
		tokenString := strings.TrimPrefix(
			authHeader,
			"Bearer ",
		)

		// Validate JWT
		claims, err := validateToken(tokenString)

		if err != nil {

			log.Println("JWT validation error:", err)

			http.Error(
				w,
				"Invalid or expired token",
				http.StatusUnauthorized,
			)

			return
		}

		// Authenticated username
		log.Println(
			"Authenticated user:",
			claims.Username,
		)

		// Authentication successful
		next(w, r)
	}
}

// =========================
// PROTECTED HANDLER
// =========================

func protectedHandler(w http.ResponseWriter, r *http.Request) {

	response := map[string]interface{}{
		"msg":    "You are authenticated",
		"status": 200,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}
func profileHandler(w http.ResponseWriter, r *http.Request){
	w.Header().set("Content_type","application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":"This should be private.but right now anyone can see it"
	})
}

// =========================
// MAIN
// =========================

func main() {
mux:=http.NewServeMux()

	// Public endpoint
	mux.HandleFunc(
		"/api/v1/health",
		healthHandler,
	)

	// Login endpoint
	mux.HandleFunc(
		"/api/v1/login",
		loginHandler,
	)

	// Protected endpoint
	mux.HandleFunc(
		"/api/v1/profile",
		authMiddleware(protectedHandler),
	)
addr:=":8080"
	log.Println(
		"Server running on",8080
	)

	if err := http.ListenAndServe(":8080", nil); err != nil {

		log.Fatal(
			"Server failed to start:",
			err,
		)
	}
}