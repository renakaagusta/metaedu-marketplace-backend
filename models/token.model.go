package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

type Token struct {
	ID                   uuid.UUID     `json:"id"`
	PreviousID           uuid.UUID     `json:"previous_id"`
	TokenIndex           int           `json:"token_index"`
	Title                string        `json:"title"`
	Description          string        `json:"description"`
	CategoryID           uuid.UUID     `json:"category_id"`
	Category             TokenCategory `json:"category"`
	CollectionID         uuid.UUID     `json:"collection_id"`
	Collection           Collection    `json:"collection"`
	Image                string        `json:"image"`
	Uri                  string        `json:"uri"`
	SourceID             uuid.UUID     `json:"source_id"`
	Source               Fraction      `json:"source"`
	FractionID           uuid.UUID     `json:"fraction_id"`
	Fraction             Fraction      `json:"fraction"`
	Supply               int           `json:"supply"`
	LastPrice            float64       `json:"last_price"`
	InitialPrice         float64       `json:"initial_price"`
	Views                int           `json:"views"`
	NumberOfTransactions int           `json:"number_of_transactions"`
	VolumeTransactions   float64       `json:"volume_transactions"`
	CreatorID            uuid.UUID     `json:"creator_id"`
	Creator              User          `json:"creator"`
	Attributes           Attributes    `json:"attributes"`
	Status               string        `json:"status"`
	TransactionHash      string        `json:"transaction_hash"`
	CreatedAt            sql.NullTime  `json:"created_at"`
	UpdatedAt            sql.NullTime  `json:"updated_at"`
}

type Attribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}

type Attributes []Attribute

func (s Attributes) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *Attributes) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	}
	return errors.New("type assertion failed")
}
