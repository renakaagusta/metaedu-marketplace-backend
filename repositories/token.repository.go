package repositories

import (
	"database/sql"
	"fmt"
	models "metaedu-marketplace/models"

	"github.com/google/uuid"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db}
}

func (r *TokenRepository) InsertToken(token models.Token) (string, error) {
	sqlStatement := `INSERT INTO tokens (
		title,
		description,
		category_id,
		collection_id,
		image,
		uri,
		fraction_id,
		quantity,
		last_price,
		updated_at,
		created_at
	  ) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
	  )
	  RETURNING id`

	var id string

	err := r.db.QueryRow(sqlStatement, token.Title, token.Description, token.CategoryID, token.CollectionID, token.Image, token.Uri, token.FractionID, token.Quantity, token.LastPrice, token.UpdatedAt, token.CreatedAt).Scan(&id)

	if err != nil {
		return id, err
	}

	return id, nil
}

func (r *TokenRepository) GetTokenList(limit int, offset int, keyword string, orderBy string, orderOption string) ([]models.Token, error) {
	var tokens []models.Token

	sqlStatement := `select id, title, description, category_id, collection_id, image, uri, fraction_id, quantity, last_price, updated_at, created_at from tokens where title LIKE '%' || $1 || '%' order by ` + orderBy + ` ` + orderOption + ` offset $2 limit $3`

	rows, err := r.db.Query(sqlStatement, keyword, offset, limit)

	if err != nil {
		return tokens, err
	}

	defer rows.Close()
	for rows.Next() {
		var token models.Token
		err = rows.Scan(&token.ID, &token.Title, &token.Description, &token.CategoryID, &token.CollectionID, &token.Image, &token.Uri, &token.FractionID, &token.Quantity, &token.LastPrice, &token.UpdatedAt, &token.CreatedAt)

		if err != nil {
			return tokens, err
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (r *TokenRepository) GetTokenData(id uuid.UUID) (models.Token, error) {
	sqlStatement := `SELECT id, title, description, category_id, collection_id, image, uri, fraction_id, quantity, updated_at, created_at FROM tokens WHERE id = $1`

	var token models.Token
	rows, err := r.db.Query(sqlStatement, id)

	if err != nil {
		return token, err
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&token.ID, &token.Title, &token.Description, &token.CategoryID, &token.CollectionID, &token.Image, &token.Uri, &token.FractionID, &token.Quantity, &token.UpdatedAt, &token.CreatedAt)

		if err != nil {
			return token, err
		}
	}

	return token, nil
}

func (r *TokenRepository) UpdateToken(id uuid.UUID, token models.Token) error {
	sqlStatement := `UPDATE tokens
	SET title = $2, description = $3, category_id = $4, collection_id = $5, image = $6, uri = $7, fraction_id = $8, quantity = $9, updated_at = $10
	WHERE id = $1;`

	result, err := r.db.Exec(sqlStatement, id, token.Title, token.Description, token.CategoryID, token.CollectionID, token.Image, token.Uri, token.FractionID, token.Quantity, token.UpdatedAt)

	fmt.Println(token.Quantity)
	fmt.Println(result)

	if err != nil {
		return err
	}

	return nil
}

func (r *TokenRepository) DeleteToken(id uuid.UUID) error {
	sqlStatement := `DELETE FROM tokens WHERE id = $1`

	_, err := r.db.Exec(sqlStatement, id)

	if err != nil {
		return err
	}

	return nil
}
