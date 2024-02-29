package itemsserviceshandlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
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

// LaundryItemsService is a interface of the request body to add items to service
type LaundryItemsService struct {
	Items []struct {
		LaundryItemID uuid.UUID `json:"laundry_item_id"`
		ItemQuantity  int       `json:"item_quantity"`
		Observation   string    `json:"observation"`
	} `json:"items"`
}

// AddItemsServicesHandler handles the creation of a laundry services
func AddItemsServicesHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newItemsService LaundryItemsService
		err := json.NewDecoder(r.Body).Decode(&newItemsService)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		vars := mux.Vars(r)
		serviceIDStr, ok := vars["serviceID"]
		if !ok {
			http.Error(w, "Service ID not provided in URL", http.StatusBadRequest)
			return
		}

		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "id", Message: "Invalid service ID", Status: http.StatusBadRequest},
				"error":   "Validation failed",
			})
			return
		}

		err = validateServiceIDExists(db, serviceIDStr)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "id", Message: err.Error(), Status: http.StatusNotFound},
				"error":   "Validation failed",
			})
			return
		}

		err = validateLaundryItemsExistence(db, newItemsService.Items)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "items", Message: err.Error(), Status: http.StatusNotFound},
				"error":   "Validation failed",
			})
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

		err = insertLaundryItems(tx, serviceIDStr, newItemsService)
		if err != nil {
			http.Error(w, "Error to insert items in service_items", http.StatusBadRequest)
			return
		}

		var isPiece bool
		err = tx.Get(&isPiece, "SELECT is_piece FROM laundry_services WHERE id=$1", serviceID)

		if err != nil {
			http.Error(w, "Error to find service in the database", http.StatusInternalServerError)
			return
		}

		if isPiece {
			calculateUpdatedTotalPrice(tx, serviceIDStr)
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newItemsService)
	}
}

func insertLaundryItems(tx *sqlx.Tx, serviceID string, itemsService LaundryItemsService) error {
	for _, item := range itemsService.Items {
		_, err := tx.Exec(`
			INSERT INTO laundry_items_services (laundry_service_id, laundry_item_id, item_quantity, observation)
			VALUES ($1, $2, $3, $4)`,
			serviceID, item.LaundryItemID, item.ItemQuantity, item.Observation)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateLaundryItemsExistence(db *sqlx.DB, items []struct {
	LaundryItemID uuid.UUID `json:"laundry_item_id"`
	ItemQuantity  int       `json:"item_quantity"`
	Observation   string    `json:"observation"`
}) error {
	for _, item := range items {
		var exists bool
		err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM laundry_items WHERE id = $1)", item.LaundryItemID)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("laundry item with ID %s does not exist", item.LaundryItemID)
		}
	}
	return nil
}

func validateServiceIDExists(db *sqlx.DB, serviceID string) error {
	var exists bool
	err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM laundry_services WHERE id = $1)", serviceID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Service with ID %s does not exist", serviceID)
	}
	return nil
}
