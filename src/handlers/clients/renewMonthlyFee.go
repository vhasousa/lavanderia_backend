package clientshandlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// Update represents the fields allowed to be updated for a client's monthly fee
type Update struct {
	RenewalDate string `json:"renewal_date"`
}

// RenewMonthlyFeeHandler handles the renewal of monthly fees for a client
func RenewMonthlyFeeHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get client ID from URL parameters
		vars := mux.Vars(r)
		clientIDStr, ok := vars["id"]
		if !ok {
			http.Error(w, "client ID not provided in URL", http.StatusBadRequest)
			return
		}

		// Convert clientIDStr to uuid.UUID
		clientID, err := uuid.Parse(clientIDStr)
		if err != nil {
			http.Error(w, "Invalid client ID", http.StatusBadRequest)
			return
		}

		// Parse request body to get the renewal information
		var update Update
		err = json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		tx, err := db.Beginx()
		if err != nil {
			http.Error(w, "Error starting database transaction", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err != nil {
				tx.Rollback()
				return
			}
			tx.Commit()
		}()

		// Update the renewal date and status in the database
		_, err = tx.Exec("UPDATE clients SET monthly_date=$1, is_mensal=$2 WHERE id=$3", update.RenewalDate, true, clientID)

		if err != nil {
			http.Error(w, "Error updating client's monthly fee renewal in the database", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}
}
