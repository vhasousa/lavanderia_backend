package serviceshandlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// DeleteServiceHandler handles the deletion of a user by ID
func DeleteServiceHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get service ID from URL parameters
		vars := mux.Vars(r)
		serviceIDStr, ok := vars["id"]
		if !ok {
			http.Error(w, "Service ID not provided in URL", http.StatusBadRequest)
			return
		}

		// Parse serviceIDStr as UUID
		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			http.Error(w, "Invalid service ID", http.StatusBadRequest)
			return
		}

		// Execute the delete query
		_, err = db.Exec("DELETE FROM laundry_services WHERE id=$1", serviceID)
		if err != nil {
			http.Error(w, "Error deleting service from the database", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}
}
