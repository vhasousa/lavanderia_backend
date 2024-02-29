package itemshandlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// DeleteItemHandler handles the deletion of a item by ID
func DeleteItemHandler(db sqlx.ExtContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get item ID from URL parameters
		vars := mux.Vars(r)
		itemIDStr, ok := vars["id"]
		if !ok {
			http.Error(w, "item ID not provided in URL", http.StatusBadRequest)
			return
		}

		// Parse itemIDStr as UUID
		itemID, err := uuid.Parse(itemIDStr)
		if err != nil {
			http.Error(w, "Invalid item ID", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		if err := validateDeleteItem(ctx, db, itemIDStr); err != nil {
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

		// Execute the delete query
		_, err = db.ExecContext(ctx, "DELETE FROM laundry_items WHERE id=$1", itemID)
		if err != nil {
			http.Error(w, "Error deleting item from database", http.StatusInternalServerError)
			return
		}
		// Return success response
		w.WriteHeader(http.StatusOK)
	}
}

func validateDeleteItem(ctx context.Context, db sqlx.ExtContext, itemID string) error {
	var exists bool
	// Fixed the query by adding a closing parenthesis
	err := sqlx.GetContext(ctx, db, &exists, "SELECT EXISTS(SELECT 1 FROM laundry_items WHERE id=$1)", itemID)
	if err != nil {
		return ValidationError{Field: "id", Message: "Failed to validate item existence", Status: http.StatusNotFound}
	}
	if !exists { // Inverted the logic here to check if the item does not exist
		return ValidationError{Field: "id", Message: "No item with this ID exists", Status: http.StatusNotFound}
	}

	var fkExists bool
	err = sqlx.GetContext(ctx, db, &fkExists, "SELECT EXISTS(SELECT 1 FROM laundry_items_services WHERE laundry_item_id=$1)", itemID)
	if err != nil {
		return ValidationError{Field: "id", Message: "Failed to check foreign key constraints", Status: http.StatusConflict}
	}
	if fkExists {
		return ValidationError{Field: "id", Message: "Cannot delete item because it is referenced in other records", Status: http.StatusConflict}
	}

	return nil
}
