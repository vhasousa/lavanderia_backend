package testhandlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver

	"lavanderia/entities"
	itemshandlers "lavanderia/handlers/items"
)

func TestCreateItemHandler(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(tx *sqlx.Tx) // Function to setup preconditions for each test case
		item       entities.LaundryItemsEntity
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "Valid Item",
			item:       entities.LaundryItemsEntity{Name: "Test Item", Price: 10.00},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name:       "Empty Name",
			item:       entities.LaundryItemsEntity{Name: "", Price: 10.00},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "Negative Price",
			item:       entities.LaundryItemsEntity{Name: "Test Item", Price: -10.00},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "Duplicate Name",
			setupFunc: func(tx *sqlx.Tx) {
				// Insert an item with the same name to setup the duplicate condition
				_, err := tx.Exec("INSERT INTO laundry_items (name, price) VALUES ($1, $2)", "Duplicate Name Item", 10.00)
				if err != nil {
					t.Fatalf("Setup failed: Unable to insert initial item for duplicate name test: %v", err)
				}
			},
			item:       entities.LaundryItemsEntity{Name: "Duplicate Name Item", Price: 20.00},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tx, err := db.Beginx() // Start a new transaction
			if err != nil {
				t.Fatalf("Failed to begin transaction: %v", err)
			}

			// Run any setup function if it exists
			if tc.setupFunc != nil {
				tc.setupFunc(tx)
			}

			handler := itemshandlers.CreateItemHandler(tx)

			itemJSON, _ := json.Marshal(tc.item)
			req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(itemJSON))
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tc.wantStatus {
				t.Errorf("Expected status code %d, got %d", tc.wantStatus, recorder.Code)
			}

			if !tc.wantErr {
				var insertedItem entities.LaundryItemsEntity
				err := tx.Get(&insertedItem, "SELECT name, price FROM laundry_items WHERE name=$1", tc.item.Name)
				if err != nil {
					t.Fatalf("Expected item to be inserted, but got error: %v", err)
				}

				if insertedItem.Name != tc.item.Name || insertedItem.Price != tc.item.Price {
					t.Errorf("Inserted item does not match: got %+v, want %+v", insertedItem, tc.item)
				}

				if err := tx.Rollback(); err != nil {
					t.Fatalf("Failed to rollback transaction: %v", err)
				}
			} else {
				var errResponse map[string]interface{}
				json.NewDecoder(recorder.Body).Decode(&errResponse)
				if errResponse["error"] == nil {
					t.Errorf("Expected error response, got none")
				}
			}
		})
	}
}
