package models

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	CategoryID   string    `json:"category_id"`
	CollectionID string    `json:"collection_id"`
	Image        string    `json:"image"`
	Uri          string    `json:"uri"`
	FractionID   string    `json:"fraction_id"`
	Quantity     int       `json:"quantity"`
	LastPrice    int       `json:"last_price"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
