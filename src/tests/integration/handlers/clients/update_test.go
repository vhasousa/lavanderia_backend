package testhandlers

import (
	"bytes"
	"encoding/json"
	clientshandlers "lavanderia/handlers/clients"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// UpdateClient is the interface of the return
type UpdateClient struct {
	ID        uuid.UUID `json:"id" db:"id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Phone     string    `json:"phone" db:"phone"`
	Username  string    `json:"username" db:"username"`
	IsMonthly bool      `json:"is_monthly" db:"is_mensal"`

	AddressID  string `json:"address_id" db:"address_id"`
	Street     string `json:"street" db:"street"`
	City       string `json:"city" db:"city"`
	State      string `json:"state" db:"state"`
	PostalCode string `json:"postal_code" db:"postal_code"`
	Number     string `json:"number" db:"number"`
	Complement string `json:"complement" db:"complement"`
	Landmark   string `json:"landmark" db:"landmark"`
}

func TestUpdateClientHandler(t *testing.T) {
	setupFunc := func(db *sqlx.DB) (string, string) {
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

		return clientID, addressID
	}

	tests := []struct {
		name       string
		setup      func(db *sqlx.DB) (string, string)
		client     UpdateClient
		wantStatus int
		wantErr    bool
	}{
		{
			name:  "Valid client update",
			setup: setupFunc,
			client: UpdateClient{
				FirstName:  "Vitor",
				LastName:   "Bastos",
				Username:   "vitor.bastos",
				Phone:      "999999999",
				IsMonthly:  false,
				AddressID:  "",
				Street:     "Avenida Geraldo Lima",
				City:       "Valença",
				State:      "RJ",
				PostalCode: "27600000",
				Number:     "900",
				Complement: "",
				Landmark:   "",
			},
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

			clientID, addressID := tc.setup(db)

			tc.client.AddressID = addressID

			handler := clientshandlers.UpdateClientHandler(db)

			clientJSON, _ := json.Marshal(tc.client)
			req, _ := http.NewRequest("PUT", "/clients"+clientID, bytes.NewBuffer(clientJSON))
			recorder := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"id": clientID})

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tc.wantStatus {
				t.Errorf("Expected status code %d, got %d", tc.wantStatus, recorder.Code)
			}
		})
	}
}
