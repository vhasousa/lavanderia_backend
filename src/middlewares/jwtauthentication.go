package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
)

// JWTAuthentication authenticates the user
func JWTAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Erro ao carregar o arquivo .env")
		}

		jwtKey := os.Getenv("JWT_KEY")

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authorizationHeader, " ")
		if len(bearerToken) != 2 {
			http.Error(w, "Invalid Authorization token", http.StatusUnauthorized)
			return
		}

		token, error := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
			// Ensure token signing method is as expected
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			// Return the key used for signing
			return []byte(jwtKey), nil
		})

		if error != nil {
			http.Error(w, error.Error(), http.StatusUnauthorized)
			return
		}

		if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Token is valid, you can add additional checks on claims here, and even set context values if needed
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Invalid Authorization token", http.StatusUnauthorized)
		}
	})
}
