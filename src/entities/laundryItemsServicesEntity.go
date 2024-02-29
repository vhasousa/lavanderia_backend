package entities

import (
	"github.com/google/uuid"
)

// LaundryItemsServicesEntity represents the laundry_items_services table in the database
type LaundryItemsServicesEntity struct {
	LaundryServiceID uuid.UUID `json:"laundry_service_id" db:"laundry_service_id"`
	LaundryItemID    uuid.UUID `json:"laundry_item_id" db:"laundry_item_id"`
	ItemQuantity     int       `json:"item_quantity" db:"item_quantity"`
	Observation      string    `json:"observation" db:"observation"`
}
