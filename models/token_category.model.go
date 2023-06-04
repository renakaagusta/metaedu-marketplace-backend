package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type TokenCategory struct {
	ID          uuid.UUID    `json:"id"`
	PreviousID  uuid.UUID    `json:"previous_id"`
	Title       string       `json:"title"`
	Icon        string       `json:"icon"`
	Description string       `json:"description"`
	Status      string       `json:"status"`
	CreatedAt   sql.NullTime `json:"created_at"`
	UpdatedAt   sql.NullTime `json:"updated_at"`
}
