package models

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID         uuid.UUID `json:"id"`
	UserFromID string    `json:"user_from_id"`
	UserToId   string    `json:"user_to_id"`
	TokenId    string    `json:"token_id"`
	Quantity   int       `json:"quantity"`
	Price      int       `json:"price"`
	GassFee    int       `json:"gass_fee"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
