package testhandlers

import (
	"fmt"
	"lavanderia/entities"
	itemsserviceshandlers "lavanderia/handlers/laundryItemsServices"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func TestDeleteItemServiceHandler(t *testing.T) {
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

	setupInitialService := func(db *sqlx.DB) map[string]string {
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

		var itemToDelete string
		err = db.Get(&itemToDelete, "SELECT id FROM laundry_items WHERE name = $1", "Item 1")

		if err != nil {
			t.Fatalf("Failed to fetch created service: %v", err)
		}

		return map[string]string{
			"serviceID": serviceID,
			"itemID":    itemToDelete,
		}
	}

	tests := []struct {
		name           string
		setupFunc      func(db *sqlx.DB) map[string]string
		wantStatus     int
		wantErr        bool
		wantTotalPrice float64
	}{
		{
			name:           "Update Item Service quantity",
			setupFunc:      setupInitialService,
			wantStatus:     http.StatusOK,
			wantErr:        false,
			wantTotalPrice: 50,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := itemsserviceshandlers.DeleteItemServiceHandler(db)

			tx, err := db.Beginx()
			if err != nil {
				t.Fatalf("Failed to begin transaction: %v", err)
			}

			ids := tc.setupFunc(db)

			serviceID := ids["serviceID"]
			itemID := ids["itemID"]

			req, _ := http.NewRequest("PATCH", fmt.Sprintf("/services/%s/items/%s", serviceID, itemID), nil)
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"serviceID": serviceID, "itemID": itemID})

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tc.wantStatus {
				t.Errorf("Expected status code %d, got %d", tc.wantStatus, recorder.Code)
			}

			if !tc.wantErr {
				var totalPrice float64
				err := db.Get(&totalPrice, "SELECT total_price FROM laundry_services WHERE id = $1", serviceID)

				if err != nil {
					t.Fatalf("Failed to fetch created service: %v", err)
				}

				if totalPrice != tc.wantTotalPrice {
					t.Errorf("Expected total price to be %.2f, got %.2f", tc.wantTotalPrice, totalPrice)
				}
			} else {
				// Verify that an appropriate error message is returned
			}

			// Rollback the transaction to ensure the test is not affecting the database
			if err := tx.Rollback(); err != nil {
				t.Fatalf("Failed to rollback transaction: %v", err)
			}
		})
	}
}
