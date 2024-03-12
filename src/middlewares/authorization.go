package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
)

// RoleAuthorization checks if the user role has access to the route
func RoleAuthorization(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := godotenv.Load()
			if err != nil {
				log.Fatal("Erro ao carregar o arquivo .env")
			}

			key := os.Getenv("JWT_KEY")

			var jwtKey = []byte(key)

			tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			claims := &jwt.MapClaims{}
			_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
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
