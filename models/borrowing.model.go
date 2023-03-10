package models

import (
	"time"

	"github.com/google/uuid"
)

type Borrowing struct {
	ID        uuid.UUID `json:"id"`
	UserID    string    `json:"user_id"`
	TokenID   string    `json:"token_id"`
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
