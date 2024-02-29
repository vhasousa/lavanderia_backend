package itemshandlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"

	"lavanderia/entities"
)

// ListItemsHandler handles the listing of all items
func ListItemsHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get page and limit from query params, with defaults
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))

		var totalPages int
		var items []entities.LaundryItemsEntity
		var response map[string]interface{}

		if page < 1 {
			err := db.Select(&items, "SELECT id, name, price FROM laundry_items ORDER BY name")

			if err != nil {
				http.Error(w, "Error retrieving users from database", http.StatusInternalServerError)
				return
			}

			response = map[string]interface{}{
				"items": items,
			}
		} else {
			limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
			if limit < 1 {
				limit = 10 // default limit
			}

			offset := (page - 1) * limit

			// Query all items from the database
			err := db.Select(&items, "SELECT id, name, price FROM laundry_items ORDER BY name LIMIT $1 OFFSET $2", limit, offset)

			if err != nil {
				http.Error(w, "Error retrieving users from database", http.StatusInternalServerError)
				return
			}

			// Count total number of items for pagination
			var totalItems int
			err = db.Get(&totalItems, "SELECT COUNT(*) FROM laundry_items")
			if err != nil {
				http.Error(w, "Error counting items", http.StatusInternalServerError)
				return
			}

			totalPages = (totalItems + limit - 1) / limit

			// Format response with clients map and pagination info
			response = map[string]interface{}{
				"items":       items,
				"page":        page,
				"total_pages": totalPages,
			}
		}

		// Return the list of users as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
