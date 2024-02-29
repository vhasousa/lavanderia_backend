package itemshandlers

import (
	"context"
	"encoding/json"
	"lavanderia/entities"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// ShowItemHandler handles the Showing the item detail
func ShowItemHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get itemID from URL parameters
		vars := mux.Vars(r)
		itemIDStr, ok := vars["id"]
		if !ok {
			http.Error(w, "item ID not provided in URL", http.StatusBadRequest)
			return
		}

		itemID, err := uuid.Parse(itemIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "id", Message: "Invalid item ID", Status: http.StatusBadRequest},
				"error":   "Validation failed",
			})
			return
		}

		ctx := r.Context()

		// Validate the input data
		if err := validateDetailItem(ctx, db, itemID); err != nil {
			if ve, ok := err.(ValidationError); ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(ve.Status)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"details": []ValidationError{ve},
					"error":   "Validation failed",
				})
				return
			}
		}

		// Query all clients from the database with pagination
		var item entities.LaundryItemsEntity
		err = db.Get(&item,
			`SELECT id, name, price
		  FROM
			laundry_items
		  WHERE id=$1
		  ORDER BY
			name`, itemID)
		if err != nil {
			http.Error(w, "Error retrieving items from database", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item)
	}
}

func validateDetailItem(ctx context.Context, db sqlx.ExtContext, itemID uuid.UUID) error {
	var exists bool
	err := sqlx.GetContext(ctx, db, &exists, "SELECT EXISTS(SELECT 1 FROM laundry_items WHERE id=$1)", itemID)
	if err != nil {
		return ValidationError{Field: "id", Message: "Failed to validate item existence", Status: http.StatusNotFound}
	}
	if !exists { // Inverted the logic here to check if the item does not exist
		return ValidationError{Field: "id", Message: "No item with this ID exists", Status: http.StatusNotFound}
	}

	return nil
}
