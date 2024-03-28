package serviceshandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"

	"lavanderia/entities"
)

// UpdateServiceHandler handles the update of service information
func UpdateServiceHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		serviceIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "id", Message: "Service ID not provided in URL", Status: http.StatusBadRequest},
				"error":   "Validation failed",
			})
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

		err = validateServiceIDExists(db, serviceID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "id", Message: err.Error(), Status: http.StatusNotFound},
				"error":   "Validation failed",
			})
			return
		}

		var updatedService entities.LaundryServicesEntity
		err = json.NewDecoder(r.Body).Decode(&updatedService)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Message: "Invalid request payload", Status: http.StatusBadRequest},
				"error":   "Validation failed",
			})
			return
		}

		// Retrieve the CreatedAt date for the service
		var createdAt time.Time
		err = db.Get(&createdAt, "SELECT created_at FROM laundry_services WHERE id=$1", serviceID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Message: "Service not found", Status: http.StatusNotFound},
				"error":   "Service retrieval failed",
			})
			return
		}

		// Validate CompletedAt date
		if updatedService.CompletedAt != nil && updatedService.CompletedAt.Before(createdAt) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "completed_at", Message: "'CompletedAt' should not be before the 'CreatedAt' date", Status: http.StatusBadRequest},
				"error":   "Validation failed",
			})
			return
		}

		// Validate EstimatedCompletionDate
		if !updatedService.EstimatedCompletionDate.IsZero() && updatedService.EstimatedCompletionDate.Before(createdAt) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "estimated_completion_date", Message: "'EstimatedCompletionDate' should not be before the 'CreatedAt' date", Status: http.StatusBadRequest},
				"error":   "Validation failed",
			})
			return
		}

		// Verify weight when is_weight is true
		if updatedService.IsWeight && updatedService.Weight <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "weight", Message: "When 'IsWeight' is true, 'Weight' should be positive number", Status: http.StatusBadRequest},
				"error":   "Validation failed",
			})
			return
		}

		// Define valid statuses
		validStatuses := map[string]bool{
			"Separado":   true,
			"Lavando":    true,
			"Secando":    true,
			"Passando":   true,
			"Finalizado": true,
		}

		// Validate the provided status
		if _, ok := validStatuses[updatedService.Status]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "status", Message: "Invalid status. Must be one of: Separado, Lavando, Secando, Passando, Finalizado", Status: http.StatusBadRequest},
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

		var totalPrice float64

		// Check if 'is_piece' has changed to 'is_weight'
		if updatedService.IsWeight && !updatedService.IsPiece {
			totalPrice = updatedService.Weight * 20
		} else {
			totalPrice, err = calculateUpdatedTotalPrice(tx, serviceID.String())

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"details": ValidationError{Field: "total_price", Message: "When 'IsWeight' is true, 'Weight' should be positive number", Status: http.StatusBadRequest},
					"error":   "Validation failed",
				})
				return
			}
		}

		if updatedService.IsMonthly {
			err = validateClientIsMensal(db, updatedService.ClientID)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"details": ValidationError{Field: "client_id", Message: err.Error(), Status: http.StatusBadRequest},
					"error":   "Validation failed",
				})
				return
			}

			totalPrice = 0
		}

		// Update service information in the database
		_, err = db.Exec(
			"UPDATE laundry_services SET status=$1, is_paid=$2, completed_at=$3, estimated_completion_date=$4, is_weight=$5, is_piece=$6, total_price=$7, weight=$8, client_id=$9 WHERE id=$10",
			updatedService.Status,
			updatedService.IsPaid,
			updatedService.CompletedAt,
			updatedService.EstimatedCompletionDate,
			updatedService.IsWeight,
			updatedService.IsPiece,
			totalPrice,
			updatedService.Weight,
			updatedService.ClientID,
			serviceID)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Message: "Error updating service in the database", Status: http.StatusInternalServerError},
				"error":   "Validation failed",
			})
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}
}

func calculateUpdatedTotalPrice(tx *sqlx.Tx, serviceID string) (float64, error) {

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
		err = tx.Get(&price, query, item.LaundryItemID)
		if err != nil {
			return 0, err
		}

		totalPrice += (price * float64(item.ItemQuantity))
	}

	return totalPrice, nil
}

func validateServiceIDExists(db *sqlx.DB, serviceID uuid.UUID) error {
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
