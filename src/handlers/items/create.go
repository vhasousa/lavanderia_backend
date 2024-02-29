package itemshandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"

	"lavanderia/entities"
)

// ValidationError is the struct for the error return
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Status  int    `json:"status"` // HTTP status code
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("%s: %s (status %d)", ve.Field, ve.Message, ve.Status)
}

// CreateItemHandler handles the creation of a new item
func CreateItemHandler(db sqlx.ExtContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var newItem entities.LaundryItemsEntity
		err := json.NewDecoder(r.Body).Decode(&newItem)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Prepare context for database operations
		ctx := r.Context()

		// Validate the input data
		if err := validateNewItem(ctx, db, newItem); err != nil {
			if ve, ok := err.(ValidationError); ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"details": []ValidationError{ve},
					"error":   "Validation failed",
				})
				return
			}
			// Handle other types of errors
		}

		// Insert the new user into the database
		if err := sqlx.GetContext(ctx, db, &newItem, "INSERT INTO laundry_items (name, price) VALUES ($1, $2) RETURNING name, price", newItem.Name, newItem.Price); err != nil {
			http.Error(w, "Error inserting item into database", http.StatusInternalServerError)
			return
		}
		if err != nil {
			http.Error(w, "Error inserting item into database", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusCreated)
	}
}

func validateNewItem(ctx context.Context, db sqlx.ExtContext, item entities.LaundryItemsEntity) error {
	if item.Name == "" {
		return ValidationError{Field: "name", Message: "Item name cannot be empty"}
	}
	if len(item.Name) > 100 { // Assuming 100 is the maximum length for a name
		return ValidationError{Field: "name", Message: "Item name exceeds maximum length of 100 characters"}
	}
	if item.Price <= 0 {
		return ValidationError{Field: "price", Message: "Item price must be positive"}
	}

	// Check for duplicate names (case-insensitive)
	var exists bool
	err := sqlx.GetContext(ctx, db, &exists, "SELECT EXISTS(SELECT 1 FROM laundry_items WHERE LOWER(name) = LOWER($1))", item.Name)
	if err != nil {
		return ValidationError{Field: "name", Message: "Failed to validate item name uniqueness"}
	}
	if exists {
		return ValidationError{Field: "name", Message: "An item with this name already exists"}
	}

	return nil
}
