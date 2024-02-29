package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"

	"lavanderia/entities"
)

// ListUsersHandler handles the listing of all users
func ListUsersHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Query all users from the database
		var users []entities.UserEntity
		err := db.Select(&users, "SELECT id, first_name, last_name  FROM users")

		if err != nil {
			http.Error(w, "Error retrieving users from database", http.StatusInternalServerError)
			return
		}

		// Return the list of users as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}
