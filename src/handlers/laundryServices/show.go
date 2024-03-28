package serviceshandlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// ServiceDetail represents a laundry service
type ServiceDetail struct {
	ID                      string        `json:"id"`
	Items                   []ServiceItem `json:"items"`
	Status                  string        `json:"status"`
	TotalPrice              float64       `json:"total_price"`
	IsMonthly               bool          `json:"is_monthly"`
	IsPaid                  bool          `json:"is_paid"`
	IsWeight                bool          `json:"is_weight"`
	IsPiece                 bool          `json:"is_piece"`
	Weight                  float64       `json:"weight"`
	ClientID                string        `json:"client_id"`
	ClientFirstName         string        `json:"client_first_name"`
	ClientLastName          string        `json:"client_last_name"`
	EstimatedCompletionDate string        `json:"estimated_completion_date"`
	CompletedAt             *string       `json:"completed_at"`
	AddressID               string        `json:"address_id"`
	Street                  string        `json:"street"`
	City                    string        `json:"city"`
	State                   string        `json:"state"`
	PostalCode              string        `json:"postal_code"`
	Number                  string        `json:"number"`
	Phone                   string        `json:"phone"`
}

// ShowServiceHandler handles the display of a single service
func ShowServiceHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get serviceID from URL parameters
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

		// Execute the SQL query with a filter for the specific serviceID
		rows, err := db.Queryx(`
		SELECT ls.id, 
			li.id as item_id, 
			li.name, 
			lis.item_quantity, 
			lis.observation, 
			ls.status, 
			ls.is_paid, 
			ls.is_weight, 
			ls.is_piece, 
			ls.weight, 
			ls.total_price, 
			cli.id, 
			cli.first_name, 
			cli.last_name, ls.estimated_completion_date, ls.completed_at, ad.address_id, ad.street, ad.city, ad.state, ad.postal_code, ad.number, cli.phone, cli.is_mensal
		FROM laundry_items_services lis
			LEFT JOIN laundry_services ls ON lis.laundry_service_id = ls.id
			LEFT JOIN laundry_items li ON lis.laundry_item_id = li.id
			LEFT JOIN clients cli ON ls.client_id = cli.id
			LEFT JOIN address ad ON cli.address_id = ad.address_id
		WHERE ls.id = $1
		`, serviceID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Initialize variables to hold the service data
		var service ServiceDetail
		service.Items = make([]ServiceItem, 0)

		// Iterate through the rows and build the service
		for rows.Next() {
			var item ServiceItem

			err := rows.Scan(
				&service.ID,
				&item.ID,
				&item.Name,
				&item.ItemQuantity,
				&item.Observation,
				&service.Status,
				&service.IsPaid,
				&service.IsWeight,
				&service.IsPiece,
				&service.Weight,
				&service.TotalPrice,
				&service.ClientID,
				&service.ClientFirstName,
				&service.ClientLastName,
				&service.EstimatedCompletionDate,
				&service.CompletedAt,
				&service.AddressID,
				&service.Street,
				&service.City,
				&service.State,
				&service.PostalCode,
				&service.Number,
				&service.Phone,
				&service.IsMonthly,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Append the item to the service
			service.Items = append(service.Items, item)
		}

		// Convert the service to JSON
		responseJSON, err := json.Marshal(map[string]*ServiceDetail{"service": &service})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set response headers and write the JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	}
}
