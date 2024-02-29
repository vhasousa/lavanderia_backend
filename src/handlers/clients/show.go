package clientshandlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// ClientDetail is the interface of the return
type ClientDetail struct {
	ID        uuid.UUID `json:"id" db:"id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Phone     string    `json:"phone" db:"phone"`
	Username  string    `json:"username" db:"username"`

	IsMonthly   bool         `json:"is_monthly" db:"is_mensal"`
	MonthlyDate sql.NullTime `json:"monthly_date" db:"monthly_date"`

	AddressID  uuid.UUID      `json:"address_id" db:"address_id"`
	Street     string         `json:"street" db:"street"`
	City       string         `json:"city" db:"city"`
	State      string         `json:"state" db:"state"`
	PostalCode string         `json:"postal_code" db:"postal_code"`
	Number     string         `json:"number" db:"number"`
	Complement sql.NullString `json:"complement" db:"complement"`
	Landmark   sql.NullString `json:"landmark" db:"landmark"`
}

// ShowClientHandler handles the Showing of all clients with pagination
func ShowClientHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get serviceID from URL parameters
		vars := mux.Vars(r)
		clientID := vars["id"]

		// Query all clients from the database with pagination
		var client ClientDetail
		err := db.Get(&client,
			`SELECT
			cli.id,
			cli.first_name,
			cli.last_name,
			cli.phone,
			cli.username,
			cli.is_mensal,
			cli.monthly_date,
			ad.address_id,
			ad.street,
			ad.city,
			ad.state,
			ad.postal_code,
			ad.number,
			ad.complement,
			ad.landmark
		  FROM
			clients cli
			LEFT JOIN address ad ON cli.address_id = ad.address_id
		   WHERE
			cli.id = $1
		  ORDER BY
			first_name,
			last_name`, clientID)
		if err != nil {
			http.Error(w, "Error retrieving clients from database", http.StatusInternalServerError)
			return
		}

		// Format response with clients map and pagination info
		//response := map[string]interface{}{
		//	"clients": clients,
		//}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(client)
	}
}
