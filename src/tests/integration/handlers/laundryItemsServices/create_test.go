package testhandlers

import (
	"bytes"
	"encoding/json"
	"lavanderia/entities"
	itemsserviceshandlers "lavanderia/handlers/laundryItemsServices"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// LaundryItemsService is a interface of the request body to add items to service
type LaundryItemsService struct {
	Items []InsertedLaundryItem `json:"items"`
}

// Estrutura auxiliar para armazenar os detalhes dos itens inseridos
type InsertedLaundryItem struct {
	LaundryItemID string `json:"laundry_item_id"`
	ItemQuantity  int    `json:"item_quantity"` // Defina a quantidade conforme necessário
	Observation   string `json:"observation"`   // Defina uma observação padrão ou ajuste conforme necessário
}

func TestCreateItemServicesHandler(t *testing.T) {
	setupItemsToAdd := func(db *sqlx.DB) []InsertedLaundryItem {
		items := []entities.LaundryItemsEntity{
			{Name: "Item 2", Price: 20.00},
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

	setupItems := func(db *sqlx.DB) []InsertedLaundryItem {
		items := []entities.LaundryItemsEntity{
			{Name: "Item 1", Price: 10.00},
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

	setupInitialServiceWeight := func(db *sqlx.DB) string {
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

	setupInitialServicePiece := func(db *sqlx.DB) string {
		clientID := setupClient(db)
		items := setupItems(db)

		// Insert a service and associate it with the client and items
		var serviceID string
		serviceInsertQuery := "INSERT INTO laundry_services (client_id, estimated_completion_date, is_weight, weight, is_piece, is_paid, status, total_price) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
		err := db.QueryRow(serviceInsertQuery, clientID, time.Now().Add(24*time.Hour), false, 0, true, false, "Separado", 10).Scan(&serviceID)
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

	tests := []struct {
		name           string
		serviceID      string
		itemService    LaundryItemsService
		wantStatus     int
		wantErr        bool
		insertedItems  []InsertedLaundryItem
		wantTotalPrice float64
	}{
		{
			name:           "Valid Item Service by weight",
			serviceID:      setupInitialServiceWeight(db),
			wantStatus:     http.StatusCreated,
			wantErr:        false,
			wantTotalPrice: 100,
		},
		{
			name:           "Valid Item Service by piece",
			serviceID:      setupInitialServicePiece(db),
			wantStatus:     http.StatusCreated,
			wantErr:        false,
			wantTotalPrice: 30,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := itemsserviceshandlers.AddItemsServicesHandler(db)

			tx, err := db.Beginx()
			if err != nil {
				t.Fatalf("Failed to begin transaction: %v", err)
			}

			tc.itemService.Items = setupItemsToAdd(db)

			itemServiceJSON, _ := json.Marshal(tc.itemService)
			req, _ := http.NewRequest("POST", "/services"+tc.serviceID+"/items", bytes.NewBuffer(itemServiceJSON))
			recorder := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"serviceID": tc.serviceID})

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
				// Verify that an appropriate error message is returned
			}

			// Rollback the transaction to ensure the test is not affecting the database
			if err := tx.Rollback(); err != nil {
				t.Fatalf("Failed to rollback transaction: %v", err)
			}
		})
	}
}
