package testhandlers

import (
	"bytes"
	"encoding/json"
	clientshandlers "lavanderia/handlers/clients"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

type CreateClient struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	FirstName   string     `json:"first_name" db:"first_name"`
	LastName    string     `json:"last_name" db:"last_name"`
	Username    string     `json:"username" db:"username"`
	Password    string     `json:"password" db:"password"`
	Phone       string     `json:"phone" db:"phone"`
	IsAdmin     bool       `json:"is_admin" db:"is_admin"`
	IsMonthly   bool       `json:"is_monthly" db:"is_mensal"`
	MonthlyDate *time.Time `json:"monthly_date" db:"monthly_date"`
	AddressID   uuid.UUID  `json:"address_id" db:"address_id"`
	Street      string     `json:"street" db:"street"`
	City        string     `json:"city" db:"city"`
	State       string     `json:"state" db:"state"`
	PostalCode  string     `json:"postal_code" db:"postal_code"`
	Number      string     `json:"number" db:"number"`
	Complement  string     `json:"complement" db:"complement"`
	Landmark    string     `json:"landmark" db:"landmark"`
}

func TestCreateClientHandler(t *testing.T) {

	tests := []struct {
		name       string
		client     CreateClient
		wantStatus int
		wantErr    bool
	}{
		{
			name: "Valid client",
			client: CreateClient{
				FirstName:  "Gabriel",
				LastName:   "Almeida de Sousa",
				Username:   "gabriel.almeida",
				Phone:      "998548386",
				IsMonthly:  false,
				Street:     "Avenida Nilo Peçanha",
				City:       "Valença",
				State:      "RJ",
				PostalCode: "27600000",
				Number:     "900",
				Complement: "",
				Landmark:   "Perto da padaria",
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// tx, err := db.Beginx()
			// if err != nil {
			// 	t.Fatalf("Failed to begin transaction: %v", err)
			// }

			handler := clientshandlers.CreateClientHandler(db)

			clientJSON, _ := json.Marshal(tc.client)
			req, _ := http.NewRequest("POST", "/clients", bytes.NewBuffer(clientJSON))
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tc.wantStatus {
				t.Errorf("Expected status code %d, got %d", tc.wantStatus, recorder.Code)
			}
		})
	}
}
