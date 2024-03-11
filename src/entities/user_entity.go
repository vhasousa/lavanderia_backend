package entities

import "github.com/google/uuid"

// UserEntity represents the user table in the database
type UserEntity struct {
	ID        uuid.UUID `json:"id" db:"id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"password" db:"password"`
	IsAdmin   bool      `json:"is_admin" db:"is_admin"`
	Role      bool      `json:"role" db:"role"`
}
