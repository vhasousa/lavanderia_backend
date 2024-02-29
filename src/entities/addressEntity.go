package entities

import (
	"github.com/google/uuid"
)

// AddressEntity represents the address table in the database
type AddressEntity struct {
	AddressID  uuid.UUID `json:"address_id" db:"address_id"`
	Street     string    `json:"street" db:"street"`
	City       string    `json:"city" db:"city"`
	State      string    `json:"state" db:"state"`
	PostalCode string    `json:"postal_code" db:"postal_code"`
	Number     string    `json:"number" db:"number"`
	Complement string    `json:"complement" db:"complement"`
	Landmark   string    `json:"landmark" db:"landmark"`
}
