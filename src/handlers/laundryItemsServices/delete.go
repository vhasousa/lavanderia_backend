package itemsserviceshandlers

import (
	"lavanderia/entities"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// DeleteItemServiceHandler handles the deletion of a user by ID
func DeleteItemServiceHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get service ID from URL parameters
		vars := mux.Vars(r)
		serviceIDStr, ok := vars["serviceID"]
		if !ok {
			http.Error(w, "Service ID not provided in URL", http.StatusBadRequest)
			return
		}

		// Get item ID from URL parameters
		itemVars := mux.Vars(r)
		itemIDStr, ok := itemVars["itemID"]
		if !ok {
			http.Error(w, "Service ID not provided in URL", http.StatusBadRequest)
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

		// Execute the delete query
		_, err = db.Exec("DELETE FROM laundry_items_services WHERE laundry_service_id=$1 AND laundry_item_id=$2", serviceID, itemID)
		if err != nil {
			http.Error(w, "Error deleting service from the database", http.StatusInternalServerError)
			return
		}

		var isPiece bool
		err = db.Get(&isPiece, "SELECT is_piece FROM laundry_services WHERE id=$1", serviceID)

		if err != nil {
			http.Error(w, "Error to find service in the database", http.StatusInternalServerError)
			return
		}

		if isPiece {
			calculateTotalPriceWhenDeleted(db, serviceID.String())
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}
}

func calculateTotalPriceWhenDeleted(db *sqlx.DB, serviceID string) error {
	var items []entities.LaundryItemsServicesEntity

	var price float64
	var totalPrice float64

	query := `
	SELECT price
	FROM laundry_items
	WHERE id = $1
	`

	err := db.Select(&items, "SELECT * FROM laundry_items_services WHERE laundry_service_id=$1", serviceID)

	for _, item := range items {
		db.Get(&price, query, item.LaundryItemID)

		totalPrice += (price * float64(item.ItemQuantity))
	}

	// Update service total_price in the database
	_, err = db.Exec(
		"UPDATE laundry_services SET total_price=$1 WHERE id=$2",
		totalPrice,
		serviceID)

	if err != nil {
		return err
	}

	return nil
}
