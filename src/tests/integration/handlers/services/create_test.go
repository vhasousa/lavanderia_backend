package testhandlers

import (
	"bytes"
	"encoding/json"
	"lavanderia/entities"
	serviceshandlers "lavanderia/handlers/laundryServices"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// LaundryService is a interface of the request body to create service
type LaundryService struct {
	EstimatedCompletionDate time.Time             `json:"estimated_completion_date"`
	CompletedAt             *time.Time            `json:"completed_at"`
	Items                   []InsertedLaundryItem `json:"items"`
	Weight                  float64               `json:"weight"`
	IsWeight                bool                  `json:"is_weight"`
	IsPiece                 bool                  `json:"is_piece"`
	ClientID                string                `json:"client_id"`
	Status                  string                `json:"status"`
	IsPaid                  bool                  `json:"is_paid"`
}

// Estrutura auxiliar para armazenar os detalhes dos itens inseridos
type InsertedLaundryItem struct {
	LaundryItemID string `json:"laundry_item_id"`
	ItemQuantity  int    `json:"item_quantity"` // Defina a quantidade conforme necessário
	Observation   string `json:"observation"`   // Defina uma observação padrão ou ajuste conforme necessário
}

func TestCreateServicesHandler(t *testing.T) {
	setupItems := func(db *sqlx.DB) []InsertedLaundryItem {
		items := []entities.LaundryItemsEntity{
			{Name: "Item 1", Price: 10.00},
			{Name: "Item 2", Price: 20.00},
			{Name: "Item 3", Price: 30.00},
		}

		insertedItems := make([]InsertedLaundryItem, 0)

		for _, item := range items {
			var id string
			err := db.QueryRow("INSERT INTO laundry_items (name, price) VALUES ($1, $2) RETURNING id", item.Name, item.Price).Scan(&id)
			if err != nil {
				t.Fatalf("Setup failed: Unable to insert items for list test: %v", err)
			}

			// Adicione os detalhes do item inserido à lista, ajuste a quantidade e observação conforme necessário
			insertedItems = append(insertedItems, InsertedLaundryItem{
				LaundryItemID: id,
				ItemQuantity:  1,
				Observation:   "No issues", // Exemplo de observação
			})
		}

		// Retorne a lista de itens inseridos com seus detalhes
		return insertedItems
	}

	setupClient := func(db *sqlx.DB) string {
		var id string

		query := "INSERT INTO clients (first_name, last_name, username, password, is_admin, phone, is_mensal, monthly_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"

		err := db.QueryRow(query, "João", "Silva", "joaosilva", "senha123", "FALSE", "11987654321", "TRUE", "2024-03-01").Scan(&id)
		if err != nil {
			t.Fatalf("Setup failed: Unable to insert initial item for delete test: %v", err)
		}
		return id
	}

	tests := []struct {
		name           string
		service        LaundryService
		wantStatus     int
		wantErr        bool
		insertedItems  []InsertedLaundryItem
		wantTotalPrice float64
		errField       string
	}{
		{
			name: "Valid Service by weight",
			service: LaundryService{
				EstimatedCompletionDate: time.Now().Add(24 * time.Hour),
				Weight:                  5.0,
				IsWeight:                true,
				IsPiece:                 false,
				IsPaid:                  false,
				Items:                   setupItems(db),
				ClientID:                setupClient(db),
			},
			wantStatus:     http.StatusCreated,
			wantErr:        false,
			wantTotalPrice: 100.0,
		},
		{
			name: "Valid Service by piece",
			service: LaundryService{
				EstimatedCompletionDate: time.Now().Add(24 * time.Hour),
				IsWeight:                false,
				IsPiece:                 true,
				IsPaid:                  false,
				Items:                   setupItems(db),
				ClientID:                setupClient(db),
			},
			wantStatus:     http.StatusCreated,
			wantErr:        false,
			wantTotalPrice: 60.0,
		},
		{
			name: "Invalid Weight",
			service: LaundryService{
				EstimatedCompletionDate: time.Now().Add(24 * time.Hour),
				Weight:                  -1.0, // Invalid weight
				IsWeight:                true,
				IsPiece:                 false,
				IsPaid:                  false,
				Items:                   setupItems(db),
				ClientID:                setupClient(db),
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
			errField:   "weight",
		},
		{
			name: "No Items for Piece Based Service",
			service: LaundryService{
				EstimatedCompletionDate: time.Now().Add(24 * time.Hour),
				IsPiece:                 true,
				IsWeight:                false,
				IsPaid:                  false,
				ClientID:                setupClient(db),
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
			errField:   "items",
		},
		{
			name: "Invalid item for service",
			service: LaundryService{
				EstimatedCompletionDate: time.Now().Add(24 * time.Hour),
				IsPiece:                 true,
				IsWeight:                false,
				IsPaid:                  false,
				ClientID:                setupClient(db),
				Items: []InsertedLaundryItem{{
					LaundryItemID: "00000000-0000-0000-0000-000000000000",
					ItemQuantity:  1,
					Observation:   "No issue",
				}},
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
			errField:   "items",
		},
		{
			name: "Client not found",
			service: LaundryService{
				EstimatedCompletionDate: time.Now().Add(24 * time.Hour),
				Weight:                  1.0,
				IsWeight:                true,
				IsPiece:                 false,
				IsPaid:                  false,
				Items:                   setupItems(db),
				ClientID:                "00000000-0000-0000-0000-000000000000",
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
			errField:   "client_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := serviceshandlers.CreateServicesHandler(db)

			serviceJSON, _ := json.Marshal(tc.service)
			req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(serviceJSON))
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tc.wantStatus {
				t.Errorf("Expected status code %d, got %d", tc.wantStatus, recorder.Code)
			}

			if !tc.wantErr {
				var responseBody map[string]interface{}
				if err := json.NewDecoder(recorder.Body).Decode(&responseBody); err != nil {
					t.Fatalf("Failed to decode response body: %v", err)
				}

				createdServiceID, ok := responseBody["id"].(string)
				if !ok {
					t.Fatalf("Service ID not found in response body")
				}

				var totalPrice float64
				err := db.Get(&totalPrice, "SELECT total_price FROM laundry_services WHERE id = $1", createdServiceID)

				if err != nil {
					t.Fatalf("Failed to fetch created service: %v", err)
				}

				if totalPrice != tc.wantTotalPrice {
					t.Errorf("Expected total price to be %.2f, got %.2f", tc.wantTotalPrice, totalPrice)
				}
			} else {
				var response struct { // Define a struct that matches your JSON response structure
					Details struct {
						Field   string `json:"field"`
						Message string `json:"message"`
						Status  int    `json:"status"`
					} `json:"details"`
					Error string `json:"error"`
				}

				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response body: %v", err)
				}

				if response.Details.Field != tc.errField {
					t.Errorf("Expected error field '%s', got '%s'", tc.errField, response.Details.Field)
				}
			}
		})
	}
}
