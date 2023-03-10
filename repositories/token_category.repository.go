package repositories

import (
	"database/sql"
	"log"
	models "metaedu-marketplace/models"
)

type TokenCategoryRepository struct {
	db *sql.DB
}

func NewTokenCategoryRepository(db *sql.DB) *TokenCategoryRepository {
	return &TokenCategoryRepository{db}
}

func (r *TokenCategoryRepository) InsertTokenCategory(token models.TokenCategory) string {
	sqlStatement := `INSERT INTO token_categories (
		title,
		description,
		updated_at
	  ) VALUES (
		$1, $2, $3
	  )
	  RETURNING id`

	var id string

	err := r.db.QueryRow(sqlStatement, token.Title, token.Description, token.UpdatedAt).Scan(&id)

	if err != nil {
		log.Fatalf("Query is cannot executed %v", err)
	}

	return id
}

func (r *TokenCategoryRepository) UpdateTokenCategory(token models.TokenCategory) error {
	sqlStatement := `UPDATE tokens
	SET title = $2, description = $3, updated_at = $4
	WHERE id = $1;`

	_, err := r.db.Exec(sqlStatement, token.ID, token.Title, token.Description, token.UpdatedAt)

	if err != nil {
		log.Fatalf("Query is cannot executed %v", err)
	}

	return err
}
