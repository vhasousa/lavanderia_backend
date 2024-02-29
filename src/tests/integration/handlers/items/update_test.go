package testhandlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver

	"lavanderia/entities"
	itemshandlers "lavanderia/handlers/items"
)

func TestUpdateItemHandler(t *testing.T) {
	setupFunc := func(db *sqlx.DB) string {
		var id string
		err := db.QueryRow("INSERT INTO laundry_items (name, price) VALUES ($1, $2) RETURNING id", "Item name", 10.00).Scan(&id)
		if err != nil {
			t.Fatalf("Setup failed: Unable to insert initial item for delete test: %v", err)
		}
		return id
	}

	tests := []struct {
		name       string
		setup      func(db *sqlx.DB) string
		item       entities.LaundryItemsEntity
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "Valid Item",
			setup:      setupFunc,
			item:       entities.LaundryItemsEntity{Name: "Updated Item", Price: 15.00},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name:       "Empty Name",
			setup:      setupFunc,
			item:       entities.LaundryItemsEntity{Name: "", Price: 10.00},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "Negative Price",
			setup:      setupFunc,
			item:       entities.LaundryItemsEntity{Name: "Test Item", Price: -10.00},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "Update to a name that already exists",
			setup:      setupFunc,
			item:       entities.LaundryItemsEntity{Name: "Updated Item", Price: -10.00},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			itemID := tc.setup(db)

			tx, err := db.Beginx() // Start a new transaction
			if err != nil {
				t.Fatalf("Failed to begin transaction: %v", err)
			}

			handler := itemshandlers.CreateItemHandler(tx)

			itemJSON, _ := json.Marshal(tc.item)
			req, _ := http.NewRequest("PUT", fmt.Sprintf("/items/%s", itemID), bytes.NewBuffer(itemJSON))
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
