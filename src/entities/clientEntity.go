package entities

import (
	"time"

	"github.com/google/uuid"
)

// ClientEntity represents the user table in the database
type ClientEntity struct {
	ID          uuid.UUID `json:"id" db:"id"`
	FirstName   string    `json:"first_name" db:"first_name"`
	LastName    string    `json:"last_name" db:"last_name"`
	Username    string    `json:"username" db:"username"`
	Password    string    `json:"password" db:"password"`
	Phone       string    `json:"phone" db:"phone"`
	IsAdmin     bool      `json:"is_admin" db:"is_admin"`
	Role        bool      `json:"role" db:"role"`
	IsMonthly   bool      `json:"is_monthly" db:"is_mensal"`
	MonthlyDate time.Time `json:"monthly_date" db:"monthly_date"`
	AddressID   string    `json:"address_id" db:"address_id"`
}
