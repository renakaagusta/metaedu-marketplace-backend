package models

import (
	"time"

	"github.com/google/uuid"
)

type Collection struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"name"`
	Description string    `json:"email"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
