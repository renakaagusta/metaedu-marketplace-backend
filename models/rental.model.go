package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type Rental struct {
	ID              uuid.UUID    `json:"id"`
	PreviousID      uuid.UUID    `json:"previous_id"`
	UserID          uuid.UUID    `json:"user_id"`
	User            User         `json:"user"`
	OwnerID         uuid.UUID    `json:"owner_id"`
	Owner           User         `json:"owner"`
	TokenID         uuid.UUID    `json:"token_id"`
	Token           Token        `json:"token"`
	OwnershipID     uuid.UUID    `json:"ownership_id"`
	Ownership       Ownership    `json:"ownership"`
	Timestamp       sql.NullTime `json:"timestamp"`
	Status          string       `json:"status"`
	TransactionHash string       `json:"transaction_hash"`
	CreatedAt       sql.NullTime `json:"created_at"`
	UpdatedAt       sql.NullTime `json:"updated_at"`
}
