package clientshandlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// DeleteClientHandler handles the deletion of a client by ID
func DeleteClientHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get client ID from URL parameters
		vars := mux.Vars(r)
		clientIDStr, ok := vars["id"]
		if !ok {
			http.Error(w, "client ID not provided in URL", http.StatusBadRequest)
			return
		}

		// Parse clientIDStr as UUID
		clientID, err := uuid.Parse(clientIDStr)
		if err != nil {
			http.Error(w, "Invalid client ID", http.StatusBadRequest)
			return
		}

		// Execute the delete query
		_, err = db.Exec("DELETE FROM clients WHERE id=$1", clientID)
		if err != nil {
			http.Error(w, "Error deleting client from the database", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}
}
