package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the required information for a user login attempt,
// including the username and password fields.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the required information returned
type LoginResponse struct {
	Token string `json:"token"`
}

// User Define a struct to represent a User in the database
type User struct {
	ID       string `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Role     string `db:"role"`
	// Include other fields as necessary
}

// ValidationError is the struct for the error return
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Status  int    `json:"status"` // HTTP status code
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("%s: %s (status %d)", ve.Field, ve.Message, ve.Status)
}

// LoginHandler handles with login
func LoginHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		key := os.Getenv("JWT_KEY")
		var jwtKey = []byte(key)

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var user User
		err = db.Get(&user, "SELECT id, username, password, role FROM users WHERE username = $1", req.Username)
		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusBadRequest)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			http.Error(w, "Invalid username or password", http.StatusBadRequest)
			return
		}

		expirationTime := time.Now().Add(1 * time.Hour)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": user.Username,
			"id":       user.ID,
			"role":     user.Role,
			"exp":      expirationTime.Unix(),
		})

		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    tokenString,
			Expires:  expirationTime,
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Logged in successfully"})
	}
}
