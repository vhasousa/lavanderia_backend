package testhandlers

import (
	clientshandlers "lavanderia/handlers/clients"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
)

func TestListClientsHandler(t *testing.T) {
	setupFunc := func(db *sqlx.DB) {
		var addressID string
		err := db.QueryRow("INSERT INTO address (street, city, state, postal_code, number, complement, landmark) VALUES ($1, $2, $3,$4, $5, $6, $7) RETURNING address_id", "Avenida Nilo Peçanha", "Valença", "RJ", "27600000", "900", "", "Perto da padaria").Scan(&addressID)
		if err != nil {
			t.Fatalf("Setup failed: Unable to insert initial address test: %v", err)
		}

		var clientID string
		err = db.QueryRow("INSERT INTO clients (first_name, last_name, username, password, is_admin, phone, is_mensal,monthly_date,address_id) VALUES ($1, $2, $3,$4, $5, $6, $7, $8, $9) RETURNING id", "Gabriel", "Almeida de Sousa", "gabriel.almeida", "senha_segura", false, "998548386", false, nil, addressID).Scan(&clientID)
		if err != nil {
			t.Fatalf("Setup failed: Unable to insert initial client test: %v", err)
		}

		return
	}

	tests := []struct {
		name       string
		setup      func(db *sqlx.DB)
		client     UpdateClient
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "Valid client detail",
			setup:      setupFunc,
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// tx, err := db.Beginx()
			// if err != nil {
			// 	t.Fatalf("Failed to begin transaction: %v", err)
			// }

			tc.setup(db)

			handler := clientshandlers.ListClientsHandler(db)

			req, _ := http.NewRequest("GET", "/clients", nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tc.wantStatus {
				t.Errorf("Expected status code %d, got %d", tc.wantStatus, recorder.Code)
			}
		})
	}
}
