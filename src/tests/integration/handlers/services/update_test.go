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

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func TestUpdateServicesHandler(t *testing.T) {
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

	setupClient := func(db *sqlx.DB) uuid.UUID {
		var id uuid.UUID

		query := "INSERT INTO clients (first_name, last_name, username, password, is_admin, phone, is_mensal, monthly_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"

		err := db.QueryRow(query, "João", "Silva", "joaosilva", "senha123", "FALSE", "11987654321", "TRUE", "2024-03-01").Scan(&id)
		if err != nil {
			t.Fatalf("Setup failed: Unable to insert initial item for delete test: %v", err)
		}
		return id
	}

	// Setup initial service and items in the database
	setupInitialWeightService := func(db *sqlx.DB) string {
		clientID := setupClient(db)
		items := setupItems(db)

		// Insert a service and associate it with the client and items
		var serviceID string
		serviceInsertQuery := "INSERT INTO laundry_services (client_id, estimated_completion_date, is_weight, weight, is_piece, is_paid, status, total_price) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
		err := db.QueryRow(serviceInsertQuery, clientID, time.Now().Add(24*time.Hour), true, 5.0, false, false, "Separado", 100).Scan(&serviceID)
		if err != nil {
			t.Fatalf("Setup failed: Unable to insert initial service: %v", err)
		}

		// Associate items with the service
		for _, item := range items {
			_, err := db.Exec("INSERT INTO laundry_items_services (laundry_service_id, laundry_item_id, item_quantity, observation) VALUES ($1, $2, $3, $4)", serviceID, item.LaundryItemID, item.ItemQuantity, item.Observation)
			if err != nil {
				t.Fatalf("Setup failed: Unable to associate item with service: %v", err)
			}
		}

		return serviceID
	}

	setupInitialPieceService := func(db *sqlx.DB) string {
		clientID := setupClient(db)
		items := setupItems(db)

		// Insert a service and associate it with the client and items
		var serviceID string
		serviceInsertQuery := "INSERT INTO laundry_services (client_id, estimated_completion_date, is_weight, weight, is_piece, is_paid, status, total_price) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
		err := db.QueryRow(serviceInsertQuery, clientID, time.Now().Add(24*time.Hour), false, 0, true, false, "Separado", 60).Scan(&serviceID)
		if err != nil {
			t.Fatalf("Setup failed: Unable to insert initial service: %v", err)
		}

		// Associate items with the service
		for _, item := range items {
			_, err := db.Exec("INSERT INTO laundry_items_services (laundry_service_id, laundry_item_id, item_quantity, observation) VALUES ($1, $2, $3, $4)", serviceID, item.LaundryItemID, item.ItemQuantity, item.Observation)
			if err != nil {
				t.Fatalf("Setup failed: Unable to associate item with service: %v", err)
			}
		}

		return serviceID
	}

	completedAtTest := time.Now().Add(48 * time.Hour)
	completedAtTestPast := time.Now().Add(-48 * time.Hour)

	tests := []struct {
		name           string
		serviceID      string
		updateData     entities.LaundryServicesEntity
		wantStatus     int
		wantErr        bool
		wantTotalPrice float64
		errField       string
	}{
		{
			name:      "Valid Update to piece",
			serviceID: setupInitialWeightService(db),
			updateData: entities.LaundryServicesEntity{
				EstimatedCompletionDate: time.Now().Add(48 * time.Hour),
				Weight:                  0,
				IsWeight:                false,
				IsPiece:                 true,
				IsPaid:                  true,
				CompletedAt:             nil,
				Status:                  "Lavando",
				ClientID:                setupClient(db),
			},
			wantStatus:     http.StatusOK,
			wantTotalPrice: 60,
			wantErr:        false,
		},
		{
			name:      "Valid Update to weight",
			serviceID: setupInitialPieceService(db),
			updateData: entities.LaundryServicesEntity{
				EstimatedCompletionDate: time.Now().Add(48 * time.Hour),
				Weight:                  2,
				IsWeight:                true,
				IsPiece:                 false,
				IsPaid:                  true,
				CompletedAt:             nil,
				Status:                  "Lavando",
				ClientID:                setupClient(db),
			},
			wantStatus:     http.StatusOK,
			wantTotalPrice: 40,
			wantErr:        false,
		},
		{
			name:      "Complete service",
			serviceID: setupInitialPieceService(db),
			updateData: entities.LaundryServicesEntity{
				EstimatedCompletionDate: time.Now().Add(48 * time.Hour),
				Weight:                  2,
				IsWeight:                true,
				IsPiece:                 false,
				IsPaid:                  true,
				CompletedAt:             &completedAtTest,
				Status:                  "Lavando",
				ClientID:                setupClient(db),
			},
			wantStatus:     http.StatusOK,
			wantTotalPrice: 40,
			wantErr:        false,
		},
		{
			name:      "Invalid status",
			serviceID: setupInitialPieceService(db),
			updateData: entities.LaundryServicesEntity{
				EstimatedCompletionDate: time.Now().Add(48 * time.Hour),
				Weight:                  2,
				IsWeight:                true,
				IsPiece:                 false,
				IsPaid:                  true,
				CompletedAt:             &completedAtTest,
				Status:                  "Status inválido",
				ClientID:                setupClient(db),
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
			errField:   "status",
		},
		{
			name:      "CompletedAt before CreatedAt",
			serviceID: setupInitialPieceService(db),
			updateData: entities.LaundryServicesEntity{
				EstimatedCompletionDate: time.Now().Add(48 * time.Hour),
				Weight:                  2,
				IsWeight:                true,
				IsPiece:                 false,
				IsPaid:                  true,
				CompletedAt:             &completedAtTestPast, // CompletedAt is set 48 hours in the past
				Status:                  "Lavando",
				ClientID:                setupClient(db),
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
			errField:   "completed_at",
		},
		{
			name:      "EstimatedCompletionDate before CreatedAt",
			serviceID: setupInitialPieceService(db),
			updateData: entities.LaundryServicesEntity{
				EstimatedCompletionDate: time.Now().Add(-48 * time.Hour),
				Weight:                  2,
				IsWeight:                true,
				IsPiece:                 false,
				IsPaid:                  true,
				CompletedAt:             nil,
				Status:                  "Lavando",
				ClientID:                setupClient(db),
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
			errField:   "estimated_completion_date",
		},
		{
			name:      "Verify weight when is_weight is true weight is positive",
			serviceID: setupInitialPieceService(db),
			updateData: entities.LaundryServicesEntity{
				EstimatedCompletionDate: time.Now().Add(48 * time.Hour),
				Weight:                  0,
				IsWeight:                true,
				IsPiece:                 false,
				IsPaid:                  true,
				CompletedAt:             nil,
				Status:                  "Lavando",
				ClientID:                setupClient(db),
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
			errField:   "weight",
		},
		{
			name:       "Service ID not found",
			serviceID:  "00000000-0000-0000-0000-000000000000",
			wantStatus: http.StatusNotFound,
			wantErr:    true,
			errField:   "id",
		},
		{
			name:       "Invalid service ID",
			serviceID:  "invalid-id",
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
			errField:   "id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := serviceshandlers.UpdateServiceHandler(db)

			updateDataJSON, _ := json.Marshal(tc.updateData)
			req, _ := http.NewRequest("PUT", "/services/"+tc.serviceID, bytes.NewBuffer(updateDataJSON))
			recorder := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"id": tc.serviceID})

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tc.wantStatus {
				t.Errorf("Expected status code %d, got %d", tc.wantStatus, recorder.Code)
			}

			if !tc.wantErr {
				var totalPrice float64
				err := db.Get(&totalPrice, "SELECT total_price FROM laundry_services WHERE id = $1", tc.serviceID)

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
