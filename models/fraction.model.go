package models

import (
	"time"

	"github.com/google/uuid"
)

type Fraction struct {
	ID              uuid.UUID `json:"id"`
	TokenParentID   string    `json:"token_parent_id"`
	TokenFractionID string    `json:"token_fraction_id"`
	Timestamp       time.Time `json:"timestamp"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
