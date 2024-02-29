package testhandlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver

	"lavanderia/entities"
	itemshandlers "lavanderia/handlers/items"
)

func TestShowItemHandler(t *testing.T) {
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
		{
			name: "Invalid ID",
			setup: func(db *sqlx.DB) string {
				return "invalid-id"
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			itemID := tc.setup(db)

			handler := itemshandlers.ShowItemHandler(db)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/items/%s", itemID), nil)
			recorder := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"id": itemID})

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tc.wantStatus {
				t.Errorf("Expected status code %d, got %d", tc.wantStatus, recorder.Code)
			}
		})
	}
}
