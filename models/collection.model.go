package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type Collection struct {
	ID                   uuid.UUID       `json:"id"`
	PreviousID           uuid.UUID       `json:"previous_id"`
	Thumbnail            sql.NullString  `json:"thumbnail"`
	Cover                sql.NullString  `json:"cover"`
	Title                sql.NullString  `json:"title"`
	Views                sql.NullInt64   `json:"views"`
	NumberOfItems        sql.NullInt64   `json:"number_of_items"`
	NumberOfTransactions sql.NullInt64   `json:"number_of_transactions"`
	VolumeTransactions   sql.NullFloat64 `json:"volume_transactions"`
	Floor                sql.NullFloat64 `json:"floor"`
	Description          sql.NullString  `json:"description"`
	CategoryID           uuid.UUID       `json:"category_id"`
	Category             TokenCategory   `json:"category"`
	CreatorID            uuid.UUID       `json:"owner_id"`
	Creator              User            `json:"creator"`
	Status               sql.NullString  `json:"status"`
	TransactionHash      sql.NullString  `json:"transaction_hash"`
	CreatedAt            sql.NullTime    `json:"created_at"`
	UpdatedAt            sql.NullTime    `json:"updated_at"`
}
