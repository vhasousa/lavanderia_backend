package serviceshandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

// ServiceItem represents an item in a laundry service
type ServiceItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Observation  string `json:"observation"`
	ItemQuantity int    `json:"item_quantity"`
}

// Service represents a laundry service including client and estimated completion date
type Service struct {
	ID                      string        `json:"id"`
	Items                   []ServiceItem `json:"items"`
	Status                  string        `json:"status"`
	TotalPrice              float64       `json:"total_price"`
	IsPaid                  bool          `json:"is_paid"`
	ClientFirstName         string        `json:"client_first_name"`
	ClientLastName          string        `json:"client_last_name"`
	EstimatedCompletionDate string        `json:"estimated_completion_date"`
}

// ListServicesHandler handles the listing of all services with pagination
func ListServicesHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse pagination parameters from query string
		pageStr := r.URL.Query().Get("page")
		pageSizeStr := r.URL.Query().Get("pageSize")
		searchTerm := r.URL.Query().Get("searchTerm")
		status := r.URL.Query().Get("status")

		// Default values for page and pageSize
		page, pageSize := 1, 10

		// Override defaults if parameters are provided
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

		// Calculate total number of records (consider adding a condition based on firstName and lastName)
		var totalRecords int
		countQuery := "SELECT COUNT(*) FROM laundry_services"
		err := db.Get(&totalRecords, countQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Calculate total number of pages
		totalPages := (totalRecords + pageSize - 1) / pageSize // Ensure rounding up

		// Validate requested page number
		if page > totalPages {
			msg := fmt.Sprintf("Requested page exceeds total pages. Total pages available: %d", totalPages)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		// Construct the WHERE clause and arguments dynamically
		whereClause := ""
		whereConditions := []string{}
		args := []interface{}{pageSize, (page - 1) * pageSize}
		if searchTerm != "" {
			whereConditions = append(whereConditions, "to_tsvector('english', cli.first_name || ' ' || cli.last_name) @@ plainto_tsquery('english', $3)")
			args = append(args, searchTerm)
		}

		if status != "" {
			if searchTerm != "" {
				// If searchTerm is also provided, status will use the next placeholder number
				whereConditions = append(whereConditions, fmt.Sprintf("ls.status = $%d", len(args)+1))
			} else {
				// If only status is provided, it will use $3
				whereConditions = append(whereConditions, "ls.status = $3")
			}
			args = append(args, status)
		}

		if len(whereConditions) > 0 {
			whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
		}

		// Construct the final query
		query := fmt.Sprintf(`
            WITH TopServices AS (
                SELECT ls.id
                FROM laundry_services ls
                JOIN clients cli ON ls.client_id = cli.id
                %s
                ORDER BY ls.created_at DESC
                LIMIT $1 OFFSET $2
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

		// Execute the SQL query with dynamic arguments
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
			"search_term": searchTerm,
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
