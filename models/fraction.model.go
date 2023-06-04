package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type Fraction struct {
	ID              uuid.UUID    `json:"id"`
	PreviousID      uuid.UUID    `json:"previous_id"`
	TokenParentID   uuid.UUID    `json:"token_parent_id"`
	TokenFractionID uuid.UUID    `json:"token_fraction_id"`
	Status          string       `json:"status"`
	TransactionHash string       `json:"transaction_hash"`
	CreatedAt       sql.NullTime `json:"created_at"`
	UpdatedAt       sql.NullTime `json:"updated_at"`
}
