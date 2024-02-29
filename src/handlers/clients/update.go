package clientshandlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// UpdateClient is the interface of the return
type UpdateClient struct {
	ID        uuid.UUID `json:"id" db:"id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Phone     string    `json:"phone" db:"phone"`
	Username  string    `json:"username" db:"username"`
	IsMonthly bool      `json:"is_monthly" db:"is_mensal"`

	AddressID  uuid.UUID `json:"address_id" db:"address_id"`
	Street     string    `json:"street" db:"street"`
	City       string    `json:"city" db:"city"`
	State      string    `json:"state" db:"state"`
	PostalCode string    `json:"postal_code" db:"postal_code"`
	Number     string    `json:"number" db:"number"`
	Complement string    `json:"complement" db:"complement"`
	Landmark   string    `json:"landmark" db:"landmark"`
}

// UpdateClientHandler handles the update of client information
func UpdateClientHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get client ID from URL parameters
		vars := mux.Vars(r)
		clientIDStr, ok := vars["id"]
		if !ok {
			http.Error(w, "client ID not provided in URL", http.StatusBadRequest)
			return
		}

		// Convert clientIDStr to int
		clientID, err := uuid.Parse(clientIDStr)
		if err != nil {
			http.Error(w, "Invalid client ID", http.StatusBadRequest)
			return
		}

		// Parse request body
		var updatedClient UpdateClient
		err = json.NewDecoder(r.Body).Decode(&updatedClient)
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

		// Update client information in the database
		_, err = tx.Exec("UPDATE address SET street=$1, city=$2, state=$3, postal_code=$4, number=$5, complement=$6, landmark=$7  WHERE address_id=$8",
			updatedClient.Street,
			updatedClient.City,
			updatedClient.State,
			updatedClient.PostalCode,
			updatedClient.Number,
			updatedClient.Complement,
			updatedClient.Landmark,
			updatedClient.AddressID)
		if err != nil {
			http.Error(w, "Error updating client address in the database", http.StatusInternalServerError)
			return
		}

		// Update client information in the database
		_, err = tx.Exec("UPDATE clients SET first_name=$1, last_name=$2, username=$3, phone=$4, is_mensal=$5  WHERE id=$6", updatedClient.FirstName, updatedClient.LastName, updatedClient.Username, updatedClient.Phone, updatedClient.IsMonthly, clientID)

		if err != nil {
			http.Error(w, "Error updating client in the database", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}
}
