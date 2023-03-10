package models

import (
	"time"

	"github.com/google/uuid"
)

type Ownership struct {
	ID               uuid.UUID `json:"id"`
	Token            string    `json:"token"`
	User             string    `json:"user"`
	Quantity         int       `json:"quantity"`
	SalePrice        int       `json:"sale_price"`
	RentCost         int       `json:"rent_cost"`
	AvailableForSale bool      `json:"available_for_sale"`
	AvailableForRent bool      `json:"available_for_rent"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
