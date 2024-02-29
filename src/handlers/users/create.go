package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"

	"lavanderia/entities"
)

// CreateUserHandler handles the creation of a new user
func CreateUserHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var newUser entities.UserEntity
		err := json.NewDecoder(r.Body).Decode(&newUser)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Insert the new user into the database

		err = db.QueryRow(
			"INSERT INTO users (first_name, last_name) VALUES ($1, $2) RETURNING id, first_name, last_name",
			newUser.FirstName, newUser.LastName,
		).Scan(&newUser.ID, &newUser.FirstName, &newUser.LastName)

		if err != nil {
			http.Error(w, "Error inserting user into database", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newUser)
	}
}
