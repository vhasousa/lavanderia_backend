package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// DeleteUserHandler handles the deletion of a user by ID
func DeleteUserHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from URL parameters
		vars := mux.Vars(r)
		userIDStr, ok := vars["id"]
		if !ok {
			http.Error(w, "User ID not provided in URL", http.StatusBadRequest)
			return
		}

		// Parse userIDStr as UUID
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Execute the delete query
		_, err = db.Exec("DELETE FROM users WHERE id=$1", userID)
		if err != nil {
			http.Error(w, "Error deleting user from the database", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}
}
