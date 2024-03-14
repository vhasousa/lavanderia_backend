package middleware

import (
	"log"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
)

// RoleAuthorization checks if the user role has access to the route
func RoleAuthorization(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := godotenv.Load()
			if err != nil {
				log.Fatal("Error loading .env file")
			}

			jwtKey := os.Getenv("JWT_KEY")

			// Retrieve the token from the cookie
			cookie, err := r.Cookie("auth_token") // Replace "YourCookieName" with the actual cookie name
			if err != nil {
				if err == http.ErrNoCookie {
					// If the cookie is not set, return an unauthorized status
					http.Error(w, "No authentication cookie", http.StatusForbidden)
					return
				}
				// For any other type of error, return a bad request status
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			// Get the JWT token string from the cookie
			tokenStr := cookie.Value

			// Parse the token with claims
			claims := &jwt.MapClaims{}
			_, err = jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtKey), nil
			})

			if err != nil {
				http.Error(w, "Invalid token", http.StatusForbidden)
				return
			}

			userRole := (*claims)["role"].(string)
			isAllowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
