package clientshandlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// CreateClient represents the creation of clients
type CreateClient struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	FirstName   string     `json:"first_name" db:"first_name"`
	LastName    string     `json:"last_name" db:"last_name"`
	Username    string     `json:"username" db:"username"`
	Password    string     `json:"password" db:"password"`
	Phone       string     `json:"phone" db:"phone"`
	IsAdmin     bool       `json:"is_admin" db:"is_admin"`
	IsMonthly   bool       `json:"is_monthly" db:"is_mensal"`
	MonthlyDate *time.Time `json:"monthly_date" db:"monthly_date"`
	AddressID   uuid.UUID  `json:"address_id" db:"address_id"`
	Street      string     `json:"street" db:"street"`
	City        string     `json:"city" db:"city"`
	State       string     `json:"state" db:"state"`
	PostalCode  string     `json:"postal_code" db:"postal_code"`
	Number      string     `json:"number" db:"number"`
	Complement  string     `json:"complement" db:"complement"`
	Landmark    string     `json:"landmark" db:"landmark"`
}

// CreateClientHandler handles the creation of a new item
func CreateClientHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var newClient CreateClient
		err := json.NewDecoder(r.Body).Decode(&newClient)
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

		var isAdmin = false

		newClient.AddressID = uuid.New()

		_, err = tx.Exec(
			"INSERT INTO address (address_id, street, city, state, postal_code, number, complement, landmark) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
			newClient.AddressID, newClient.Street, newClient.City, newClient.State, newClient.PostalCode, newClient.Number, newClient.Complement, newClient.Landmark,
		)
		if err != nil {
			http.Error(w, "Error inserting address into database", http.StatusInternalServerError)
			return
		}

		if newClient.IsMonthly == false {
			newClient.MonthlyDate = nil
		}

		// Insert the new user into the database
		err = tx.QueryRow(
			"INSERT INTO clients (first_name, last_name, username, is_admin, phone, is_mensal, address_id, monthly_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING first_name, last_name",
			newClient.FirstName, newClient.LastName, newClient.Username, isAdmin, newClient.Phone, newClient.IsMonthly, newClient.AddressID, newClient.MonthlyDate,
		).Scan(&newClient.FirstName, &newClient.LastName)

		if err != nil {
			http.Error(w, "Error inserting client into database", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusCreated)
	}
}
