package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
)

// AuthStatusResponse represents the structure of the response for the auth status check.
type AuthStatusResponse struct {
	IsAuthenticated bool   `json:"isAuthenticated"`
	Username        string `json:"username,omitempty"`
	Role            string `json:"role,omitempty"`
	UserID          string `json:"userID,omitempty"`
}

// AuthStatusHandler checks the user's authentication status based on the JWT token in the HTTP-Only cookie.
func AuthStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the "auth_token" cookie
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, the user is not authenticated
			json.NewEncoder(w).Encode(AuthStatusResponse{IsAuthenticated: false})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the JWT string from the cookie
	tokenStr := cookie.Value

	// Initialize a new instance of `Claims`
	claims := &jwt.MapClaims{}

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	key := os.Getenv("JWT_KEY")
	var jwtKey = []byte(key)

	// Parse the JWT string and store the result in `claims`
	tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if the token is expired
	if exp, ok := (*claims)["exp"].(float64); ok {
		now := time.Now().Unix()
		if int64(exp) < now {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// At this point, the user is authenticated
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthStatusResponse{
		IsAuthenticated: true,
		Username:        (*claims)["username"].(string),
		Role:            (*claims)["role"].(string),
		UserID:          (*claims)["id"].(string),
	})
}
