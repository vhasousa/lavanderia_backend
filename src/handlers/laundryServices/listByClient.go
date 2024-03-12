package serviceshandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// ListServicesByClientHandler handles the listing of all services with pagination
func ListServicesByClientHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		clientIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "id", Message: "client ID not provided in URL", Status: http.StatusBadRequest},
				"error":   "Validation failed",
			})
			return
		}

		clientID, err := uuid.Parse(clientIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"details": ValidationError{Field: "id", Message: "Invalid client ID", Status: http.StatusBadRequest},
				"error":   "Validation failed",
			})
			return
		}

		fmt.Print(clientID)

		pageStr := r.URL.Query().Get("page")
		pageSizeStr := r.URL.Query().Get("pageSize")
		status := r.URL.Query().Get("status")

		page, pageSize := 1, 10

		if pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil {
				page = p
			}
		}
		if pageSizeStr != "" {
			if ps, err := strconv.Atoi(pageSizeStr); err == nil {
				pageSize = ps
			}
		}

		var totalRecords int
		countQuery := "SELECT COUNT(*) FROM laundry_services WHERE client_id = $1"
		err = db.Get(&totalRecords, countQuery, clientID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		totalPages := (totalRecords + pageSize - 1) / pageSize

		if page > totalPages {
			msg := fmt.Sprintf("Requested page exceeds total pages. Total pages available: %d", totalPages)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		whereClause := "WHERE ls.client_id = $1"
		args := []interface{}{clientID, pageSize, (page - 1) * pageSize}

		if status != "" {

			whereClause += fmt.Sprintf(" AND ls.status = $%d", len(args)+1)
			args = append(args, status)
		}

		query := fmt.Sprintf(`
            WITH TopServices AS (
                SELECT ls.id
                FROM laundry_services ls
                %s
                ORDER BY ls.created_at DESC
                LIMIT $2 OFFSET $3
            )
            SELECT ls.id, li.id as item_id, li.name, lis.item_quantity, lis.observation, ls.status, ls.is_paid, ls.total_price,
                   cli.first_name AS client_first_name, cli.last_name AS client_last_name, ls.estimated_completion_date
            FROM laundry_items_services lis
            JOIN TopServices ON lis.laundry_service_id = TopServices.id
            LEFT JOIN laundry_services ls ON lis.laundry_service_id = ls.id
            LEFT JOIN laundry_items li ON lis.laundry_item_id = li.id
            LEFT JOIN clients cli ON ls.client_id = cli.id
            ORDER BY ls.created_at DESC
        `, whereClause)

		rows, err := db.Queryx(query, args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Map to store services and their items
		serviceMap := make(map[string]*Service)

		// Slice to keep track of the order of service IDs
		var orderedServiceIDs []string

		// Iterate through the rows and build the result
		for rows.Next() {
			var serviceID string
			var item ServiceItem
			var service Service

			err := rows.Scan(
				&serviceID,
				&item.ID,
				&item.Name,
				&item.ItemQuantity,
				&item.Observation,
				&service.Status,
				&service.IsPaid,
				&service.TotalPrice,
				&service.ClientFirstName,
				&service.ClientLastName,
				&service.EstimatedCompletionDate)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Check if the service already exists in the map
			if existingService, ok := serviceMap[serviceID]; ok {
				// Service exists, append the item to its items
				existingService.Items = append(existingService.Items, item)
			} else {
				// Service doesn't exist, create a new service with the item
				service.ID = serviceID
				service.Items = []ServiceItem{item}
				serviceMap[serviceID] = &service
				// Record the order of this new service ID
				orderedServiceIDs = append(orderedServiceIDs, serviceID)
			}
		}

		// Use the orderedServiceIDs slice to extract services in the correct order
		var result []Service
		for _, serviceID := range orderedServiceIDs {
			if service, ok := serviceMap[serviceID]; ok {
				result = append(result, *service)
			}
		}

		// Convert the result to JSON
		responseJSON, err := json.Marshal(map[string]interface{}{
			"services":    result,
			"page":        page,
			"total_pages": totalPages,
		})
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
