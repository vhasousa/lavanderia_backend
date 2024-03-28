package entities

import (
	"time"

	"github.com/google/uuid"
)

// LaundryServicesEntity represents the laundry_services table in the database
type LaundryServicesEntity struct {
	ID                      uuid.UUID  `json:"id" db:"id"`
	Status                  string     `json:"status" db:"status"`
	Type                    string     `json:"type" db:"type"`
	CreatedAt               time.Time  `json:"created_at" db:"created_at"`
	CompletedAt             *time.Time `json:"completed_at" db:"completed_at"`
	EstimatedCompletionDate time.Time  `json:"estimated_completion_date" db:"estimated_completion_date"`
	TotalPrice              float64    `json:"price" db:"price"`
	Weight                  float64    `json:"weight"`
	IsWeight                bool       `json:"is_weight"`
	IsPiece                 bool       `json:"is_piece"`
	IsMonthly               bool       `json:"is_monthly"`
	ClientID                uuid.UUID  `json:"client_id"`
	IsPaid                  bool       `json:"is_paid"`
}
