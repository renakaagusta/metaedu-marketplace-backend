package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID    `json:"id"`
	PreviousID uuid.UUID    `json:"previous_id"`
	Name       string       `json:"name"`
	Email      string       `json:"email"`
	Photo      string       `json:"photo"`
	Cover      string       `json:"cover"`
	Verified   bool         `json:"verified"`
	Role       string       `json:"role"`
	Address    string       `json:"address"`
	Nonce      string       `json:"nonce"`
	Status     string       `json:"status"`
	CreatedAt  sql.NullTime `json:"created_at"`
	UpdatedAt  sql.NullTime `json:"updated_at"`
}
