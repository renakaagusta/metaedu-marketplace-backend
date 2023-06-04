package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type Ownership struct {
	ID               uuid.UUID    `json:"id"`
	PreviousID       uuid.UUID    `json:"previous_id"`
	TokenID          uuid.UUID    `json:"token_id"`
	Token            Token        `json:"token"`
	UserID           uuid.UUID    `json:"user_id"`
	User             User         `json:"user"`
	Quantity         int          `json:"quantity"`
	SalePrice        float64      `json:"sale_price"`
	RentCost         float64      `json:"rent_cost"`
	AvailableForSale bool         `json:"available_for_sale"`
	AvailableForRent bool         `json:"available_for_rent"`
	Status           string       `json:"status"`
	TransactionHash  string       `json:"transaction_hash"`
	CreatedAt        sql.NullTime `json:"created_at"`
	UpdatedAt        sql.NullTime `json:"updated_at"`
}
