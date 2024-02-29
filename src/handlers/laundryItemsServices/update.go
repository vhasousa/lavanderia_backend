package itemsserviceshandlers

import (
	"encoding/json"
	"fmt"
	"lavanderia/entities"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// LaundryItemsServiceUpdate is a interface of the request body to update items
type LaundryItemsServiceUpdate struct {
	ItemQuantity int `json:"item_quantity"`
}

// UpdateItemServiceHandler handles the update of items in the service
func UpdateItemServiceHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get service ID from URL parameters
		vars := mux.Vars(r)
		serviceIDStr, ok := vars["serviceID"]
		if !ok {
			http.Error(w, "Service ID not provided in URL", http.StatusBadRequest)
			return
		}

		err := validateServiceIDExists(db, serviceIDStr)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "serviceID", Message: err.Error(), Status: http.StatusNotFound},
				"error":   "Validation failed",
			})
			return
		}

		// Get item ID from URL parameters
		itemVars := mux.Vars(r)
		itemIDStr, ok := itemVars["itemID"]
		if !ok {
			http.Error(w, "Service ID not provided in URL", http.StatusBadRequest)
			return
		}

		err = validateItemServiceIDExists(db, itemIDStr)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "itemServiceID", Message: err.Error(), Status: http.StatusNotFound},
				"error":   "Validation failed",
			})
			return
		}

		// Parse serviceIDStr as UUID
		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			http.Error(w, "Invalid service ID", http.StatusBadRequest)
			return
		}

		// Parse itemIDStr as UUID
		itemID, err := uuid.Parse(itemIDStr)
		if err != nil {
			http.Error(w, "Invalid item ID", http.StatusBadRequest)
			return
		}

		// Parse request body
		var updatedItemService LaundryItemsServiceUpdate
		err = json.NewDecoder(r.Body).Decode(&updatedItemService)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Start a transaction
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

		// Update service information in the database
		_, err = tx.Exec("UPDATE laundry_items_services SET item_quantity=$1 WHERE laundry_service_id=$2 AND laundry_item_id=$3", updatedItemService.ItemQuantity, serviceID, itemID)

		if err != nil {
			http.Error(w, "Error updating service in the database", http.StatusInternalServerError)
			return
		}

		var isPiece bool
		err = tx.Get(&isPiece, "SELECT is_piece FROM laundry_services WHERE id=$1", serviceID)

		if err != nil {
			http.Error(w, "Error to find service in the database", http.StatusInternalServerError)
			return
		}

		if isPiece {
			calculateUpdatedTotalPrice(tx, serviceID.String())
		}
		// Return success response
		w.WriteHeader(http.StatusOK)
	}
}

func calculateUpdatedTotalPrice(tx *sqlx.Tx, serviceID string) error {
	var items []entities.LaundryItemsServicesEntity

	var price float64
	var totalPrice float64

	query := `
	SELECT price
	FROM laundry_items
	WHERE id = $1
	`

	err := tx.Select(&items, "SELECT * FROM laundry_items_services WHERE laundry_service_id=$1", serviceID)

	for _, item := range items {
		tx.Get(&price, query, item.LaundryItemID)

		totalPrice += (price * float64(item.ItemQuantity))
	}

	// Update service total_price in the database
	_, err = tx.Exec(
		"UPDATE laundry_services SET total_price=$1 WHERE id=$2",
		totalPrice,
		serviceID)

	if err != nil {
		return err
	}

	return nil
}

func validateItemServiceIDExists(db *sqlx.DB, itemServiceID string) error {
	var exists bool
	err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM laundry_items_services WHERE laundry_item_id = $1)", itemServiceID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Service with ID %s does not exist", itemServiceID)
	}
	return nil
}
