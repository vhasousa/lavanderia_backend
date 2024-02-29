package testhandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver

	itemshandlers "lavanderia/handlers/items"
)

func TestDeleteItemHandler(t *testing.T) {
	setupFunc := func(db *sqlx.DB) string {
		var id string
		err := db.QueryRow("INSERT INTO laundry_items (name, price) VALUES ($1, $2) RETURNING id", "Item to Delete", 10.00).Scan(&id)
		if err != nil {
			t.Fatalf("Setup failed: Unable to insert initial item for delete test: %v", err)
		}
		return id
	}

	tests := []struct {
		name       string
		setup      func(db *sqlx.DB) string
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "Delete Existing Item",
			setup:      setupFunc,
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "Delete Non-Existing Item",
			setup: func(db *sqlx.DB) string {
				return "00000000-0000-0000-0000-000000000000"
			},
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			itemID := tc.setup(db) // Setup and get item ID

			handler := itemshandlers.DeleteItemHandler(db)
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/items/%s", itemID), nil)
			req = mux.SetURLVars(req, map[string]string{"id": itemID})

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if recorder.Code != tc.wantStatus {
				t.Errorf("Expected status code %d, got %d", tc.wantStatus, recorder.Code)
			}

			if !tc.wantErr {
				var count int
				err := db.Get(&count, "SELECT COUNT(*) FROM laundry_items WHERE id=$1", itemID)
				if err != nil || count > 0 {
					t.Errorf("Expected item to be deleted, but it still exists or query failed: %v", err)
				}
			} else {
				var errResponse map[string]interface{}
				if err := json.NewDecoder(recorder.Body).Decode(&errResponse); err == nil {
					if errResponse["error"] == nil {
						t.Errorf("Expected error response, got none")
					}
				} else {
					t.Errorf("Failed to decode error response: %v", err)
				}
			}
		})
	}
}
