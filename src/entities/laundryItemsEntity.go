package entities

import (
	"github.com/google/uuid"
)

// LaundryItemsEntity represents the laundry_items table in the database
type LaundryItemsEntity struct {
	ID    uuid.UUID `json:"id" db:"id"`
	Name  string    `json:"name" db:"name"`
	Price float64   `json:"price" db:"price"`
}
