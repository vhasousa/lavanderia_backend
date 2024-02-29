package testhandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver

	"lavanderia/entities"
	itemshandlers "lavanderia/handlers/items"
)

func TestListItemsHandler(t *testing.T) {
	setupFunc := func(db *sqlx.DB) {
		items := []entities.LaundryItemsEntity{
			{Name: "Item 1", Price: 10.00},
			{Name: "Item 2", Price: 20.00},
			{Name: "Item 3", Price: 30.00},
		}

		for _, item := range items {
			_, err := db.Exec("INSERT INTO laundry_items (name, price) VALUES ($1, $2)", item.Name, item.Price)
			if err != nil {
				t.Fatalf("Setup failed: Unable to insert items for list test: %v", err)
			}
		}
	}

	tests := []struct {
		name       string
		setup      func(db *sqlx.DB)
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "List All Items",
			setup:      setupFunc,
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(db)

			handler := itemshandlers.ListItemsHandler(db)

			req, _ := http.NewRequest("GET", "/items", nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tc.wantStatus {
				t.Errorf("Expected status code %d, got %d", tc.wantStatus, recorder.Code)
			}

			if !tc.wantErr {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Expected valid JSON response, got error: %v", err)
				}

				items, ok := response["items"].([]interface{})
				if !ok {
					t.Errorf("Expected items to be a slice, got %T", response["items"])
				}

				expectedNumItems := 3 // Ajuste conforme a configuração do seu teste
				if len(items) != expectedNumItems {
					t.Errorf("Expected %d items, got %d", expectedNumItems, len(items))
				}

			}
		})
	}
}
