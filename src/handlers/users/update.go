package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"

	"lavanderia/entities"
)

// UpdateUserHandler handles the update of user information
func UpdateUserHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from URL parameters
		vars := mux.Vars(r)
		userIDStr, ok := vars["id"]
		if !ok {
			http.Error(w, "User ID not provided in URL", http.StatusBadRequest)
			return
		}

		// Convert userIDStr to int
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Parse request body
		var updatedUser entities.UserEntity
		err = json.NewDecoder(r.Body).Decode(&updatedUser)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Update user information in the database
		_, err = db.Exec("UPDATE users SET first_name=$1, last_name=$2 WHERE id=$3", updatedUser.FirstName, updatedUser.LastName, userID)
		if err != nil {
			http.Error(w, "Error updating user in the database", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}
}
