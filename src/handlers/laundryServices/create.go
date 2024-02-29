package serviceshandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
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

// LaundryService is a interface of the request body to create service
type LaundryService struct {
	ID                      string    `json:"id"`
	EstimatedCompletionDate time.Time `json:"estimated_completion_date"`
	Items                   []struct {
		LaundryItemID uuid.UUID `json:"laundry_item_id"`
		ItemQuantity  int       `json:"item_quantity"`
		Observation   string    `json:"observation"`
	} `json:"items"`
	Weight   float64   `json:"weight"`
	IsWeight bool      `json:"is_weight"`
	IsPiece  bool      `json:"is_piece"`
	ClientID uuid.UUID `json:"client_id"`
	IsPaid   bool      `json:"is_paid"`
}

// CreateServicesHandler handles the creation of a laundry services
func CreateServicesHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newService LaundryService
		err := json.NewDecoder(r.Body).Decode(&newService)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Message: "Invalid request payload", Status: http.StatusBadRequest},
				"error":   "Validation failed",
			})
			return
		}

		if newService.IsWeight && newService.Weight <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "weight", Message: "When 'IsWeight' is true, 'Weight' should be positive number", Status: http.StatusBadRequest},
				"error":   "Validation failed",
			})
			return
		}

		if newService.IsPiece {
			if len(newService.Items) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"details": ValidationError{Field: "items", Message: "When 'IsPiece' is true, at least one item must be specified", Status: http.StatusBadRequest},
					"error":   "Validation failed",
				})
				return
			}

			for _, item := range newService.Items {
				if item.ItemQuantity <= 0 {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"details": ValidationError{Field: "item_quantity", Message: "Each item's quantity must be a positive number when 'IsPiece' is true", Status: http.StatusBadRequest},
						"error":   "Validation failed",
					})
					return
				}
			}
		}

		err = validateLaundryItemsExistence(db, newService.Items)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "items", Message: err.Error(), Status: http.StatusNotFound},
				"error":   "Validation failed",
			})
			return
		}

		err = validateClientIDExists(db, newService.ClientID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "client_id", Message: err.Error(), Status: http.StatusNotFound},
				"error":   "Validation failed",
			})
			return
		}

		err = validateEstimatedCompletionDate(newService.EstimatedCompletionDate)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "estimated_completion_date", Message: err.Error(), Status: http.StatusBadRequest},
				"error":   "Validation failed",
			})
			return
		}

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

		newService.ID = uuid.New().String()

		serviceTotalPrice, err := calculateTotalPrice(db, newService, w)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "total_price", Message: "Error calculanting total price", Status: http.StatusInternalServerError},
				"error":   "Validation failed",
			})
			return
		}

		err = insertLaundryService(tx, newService, newService.ID, serviceTotalPrice)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Message: "Error inserting laundry service into database", Status: http.StatusInternalServerError},
				"error":   "Validation failed",
			})
			return
		}

		err = insertLaundryItems(tx, newService.ID, newService)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Message: "Error inserting laundry service items into database", Status: http.StatusInternalServerError},
				"error":   "Validation failed",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newService)
	}
}

func insertLaundryService(tx *sqlx.Tx, service LaundryService, serviceID string, totalPrice float64) error {
	var serviceStatus = "Separado"

	_, err := tx.Exec(`
		INSERT INTO laundry_services (id, status, estimated_completion_date, total_price, weight, is_weight, is_piece, client_id, is_paid)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		serviceID, serviceStatus, service.EstimatedCompletionDate, totalPrice, service.Weight, service.IsWeight, service.IsPiece, service.ClientID, service.IsPaid)

	return err
}

func insertLaundryItems(tx *sqlx.Tx, serviceID string, service LaundryService) error {
	for _, item := range service.Items {
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

func calculateTotalPrice(db *sqlx.DB, service LaundryService, w http.ResponseWriter) (float64, error) {
	var price float64
	var totalPrice float64

	// Consulta SQL que junta as tabelas e seleciona o Price com base no LaundryItemID
	query := `
        SELECT price
        FROM laundry_items
        WHERE id = $1
    `

	if service.IsPiece {
		if len(service.Items) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Message: "When 'IsPiece' is true, 'Items' should not be empty", Status: http.StatusBadRequest},
				"error":   "Validation failed",
			})
			return 0, nil
		}

		for _, item := range service.Items {
			if item.ItemQuantity <= 0 {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"details": ValidationError{Message: "Item quantities should be positive numbers", Status: http.StatusBadRequest},
					"error":   "Validation failed",
				})
				return 0, nil
			}
			err := db.Get(&price, query, item.LaundryItemID)
			if err != nil {
				return 0, err
			}

			totalPrice += (price * float64(item.ItemQuantity))
		}
	} else if service.IsWeight {
		totalPrice = service.Weight * 20
	} else {
		totalPrice = 0
	}

	return totalPrice, nil
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

func validateClientIDExists(db *sqlx.DB, clientID uuid.UUID) error {
	var exists bool
	err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM clients WHERE id = $1)", clientID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("client with ID %s does not exist", clientID)
	}
	return nil
}

func validateEstimatedCompletionDate(date time.Time) error {
	currentTime := time.Now()

	// Check if the date is in the past
	if date.Before(currentTime) {
		return fmt.Errorf("estimated completion date %s is in the past", date.Format("2006-01-02"))
	}

	// Assuming a threshold of 30 days in the future for example
	maxAllowedDate := currentTime.AddDate(0, 0, 30)
	if date.After(maxAllowedDate) {
		return fmt.Errorf("estimated completion date %s is beyond the acceptable threshold of 30 days", date.Format("2006-01-02"))
	}

	return nil
}
