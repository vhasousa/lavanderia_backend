package clientshandlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ClientList is the interface of the return
type ClientList struct {
	ID        uuid.UUID `json:"id" db:"id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Phone     string    `json:"phone" db:"phone"`
}

// ListClientsHandler handles the listing of all clients with pagination
func ListClientsHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get page and limit from query params, with defaults
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))

		var clients []ClientList
		var response map[string]interface{}

		if page < 1 {
			err := db.Select(&clients, "SELECT id, first_name, last_name, phone FROM clients ORDER BY first_name, last_name")
			if err != nil {
				http.Error(w, "Error retrieving clients from database", http.StatusInternalServerError)
				return
			}

			response = map[string]interface{}{
				"clients": clients,
			}
		} else {
			limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
			if limit < 1 {
				limit = 10 // default limit
			}

			offset := (page - 1) * limit

			// Query all clients from the database with pagination
			err := db.Select(&clients, "SELECT id, first_name, last_name, phone FROM clients ORDER BY first_name, last_name LIMIT $1 OFFSET $2", limit, offset)
			if err != nil {
				http.Error(w, "Error retrieving clients from database", http.StatusInternalServerError)
				return
			}

			// Count total number of clients for pagination
			var totalClients int
			err = db.Get(&totalClients, "SELECT COUNT(*) FROM clients")
			if err != nil {
				http.Error(w, "Error counting clients", http.StatusInternalServerError)
				return
			}

			totalPages := (totalClients + limit - 1) / limit

			// Format response with clients map and pagination info
			response = map[string]interface{}{
				"clients":     clients,
				"page":        page,
				"total_pages": totalPages,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
