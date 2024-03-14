package itemshandlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"

	"lavanderia/entities"
)

// UpdateItemHandler handles the update of item information
func UpdateItemHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get item ID from URL parameters
		vars := mux.Vars(r)
		itemIDStr, ok := vars["id"]
		if !ok {
			http.Error(w, "Item ID not provided in URL", http.StatusBadRequest)
			return
		}

		// Convert itemIDStr to int
		itemID, err := uuid.Parse(itemIDStr)
		if err != nil {
			http.Error(w, "Invalid item ID", http.StatusBadRequest)
			return
		}

		// Parse request body
		var updatedItem entities.LaundryItemsEntity
		err = json.NewDecoder(r.Body).Decode(&updatedItem)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		// Validate the input data
		if err := validateUpdateItem(ctx, db, updatedItem, itemID); err != nil {
			if ve, ok := err.(ValidationError); ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"details": []ValidationError{ve},
					"error":   "Validation failed",
				})
				return
			}
		}

		// Update item information in the database
		_, err = db.Exec("UPDATE laundry_items SET name=$1, price=$2 WHERE id=$3", updatedItem.Name, updatedItem.Price, itemID)
		if err != nil {
			http.Error(w, "Error updating item in the database", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}
}

func validateUpdateItem(ctx context.Context, db sqlx.ExtContext, item entities.LaundryItemsEntity, itemID uuid.UUID) error {
	if item.Name == "" {
		return ValidationError{Field: "name", Message: "Item name cannot be empty", Status: http.StatusBadRequest}
	}
	if len(item.Name) > 100 { // Assuming 100 is the maximum length for a name
		return ValidationError{Field: "name", Message: "Item name exceeds maximum length of 100 characters", Status: http.StatusBadRequest}
	}
	if item.Price <= 0 {
		return ValidationError{Field: "price", Message: "Item price must be positive", Status: http.StatusBadRequest}
	}

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
