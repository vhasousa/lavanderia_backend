package testhandlers

import (
	"lavanderia/entities"
	serviceshandlers "lavanderia/handlers/laundryServices"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func TestListServicesHandler(t *testing.T) {
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

	setupAddress := func(db *sqlx.DB) string {
		var id string

		query := "INSERT INTO address (street, city, state, postal_code, number, complement, landmark) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING address_id"

		err := db.QueryRow(query, "Your Street", "Your City", "YS", "Your Postal Code", "300", "Your Complement", "Your Landmark").Scan(&id)
		if err != nil {
			t.Fatalf("Setup failed: Unable to insert client address for delete test: %v", err)
		}
		return id
	}

	setupClient := func(db *sqlx.DB) string {
		addressID := setupAddress(db)
		var id string

		query := "INSERT INTO clients (first_name, last_name, username, password, is_admin, phone, is_mensal, monthly_date, address_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id"

		err := db.QueryRow(query, "João", "Silva", "joaosilva", "senha123", "FALSE", "11987654321", "TRUE", "2024-03-01", addressID).Scan(&id)
		if err != nil {
			t.Fatalf("Setup failed: Unable to insert initial client for delete test: %v", err)
		}
		return id
	}

	// Setup initial service and items in the database
	setupInitialService := func(db *sqlx.DB) error {
		clientID := setupClient(db)
		items := setupItems(db)

		// Insert a service and associate it with the client and items
		var serviceID string
		serviceInsertQuery := "INSERT INTO laundry_services (client_id, estimated_completion_date, is_weight, weight, is_piece, is_paid, status, total_price) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
		err := db.QueryRow(serviceInsertQuery, clientID, time.Now().Add(24*time.Hour), true, 5.0, false, false, "Separado", 60).Scan(&serviceID)
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

		return err
	}

	tests := []struct {
		name       string
		serviceID  string
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "Valid list",
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := serviceshandlers.ListServicesHandler(db)

			tx, err := db.Beginx()
			if err != nil {
				t.Fatalf("Failed to begin transaction: %v", err)
			}

			err = setupInitialService(db)
			if err != nil {
				t.Errorf("Error to setup initial service")
			}

			req, _ := http.NewRequest("GET", "/services", nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tc.wantStatus {
				t.Errorf("Expected status code %d, got %d", tc.wantStatus, recorder.Code)
			}

			if !tc.wantErr {
				// Verify the service and its items were updated correctly
				// This could involve querying the service by its ID and checking its fields, as well as associated items
			} else {
				// Verify that an appropriate error message is returned
			}

			if err := tx.Rollback(); err != nil {
				t.Fatalf("Failed to rollback transaction: %v", err)
			}
		})
	}
}
