package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type Transaction struct {
	ID              uuid.UUID    `json:"id"`
	PreviousID      uuid.UUID    `json:"previous_id"`
	UserFromID      uuid.UUID    `json:"user_from_id"`
	UserFrom        User         `json:"user_from"`
	UserToID        uuid.UUID    `json:"user_to_id"`
	UserTo          User         `json:"user_to"`
	OwnershipID     uuid.UUID    `json:"ownership_id"`
	Ownership       Ownership    `json:"ownership"`
	RentalID        uuid.UUID    `json:"rental_id"`
	Rental          Rental       `json:"rental"`
	TokenID         uuid.UUID    `json:"token_id"`
	Token           Token        `json:"token"`
	CollectionID    uuid.UUID    `json:"collection_id"`
	Collection      Collection   `json:"collection"`
	Type            string       `string:"type"`
	Quantity        int          `json:"quantity"`
	Amount          float64      `json:"amount"`
	GasFee          float64      `json:"gas_fee"`
	Status          string       `json:"status"`
	TransactionHash string       `json:"transaction_hash"`
	CreatedAt       sql.NullTime `json:"created_at"`
	UpdatedAt       sql.NullTime `json:"updated_at"`
}
