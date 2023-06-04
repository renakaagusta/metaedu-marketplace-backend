package repositories

import (
	"database/sql"
	models "metaedu-marketplace/models"

	"github.com/google/uuid"
)

type TokenCategoryRepository struct {
	db *sql.DB
}

func NewTokenCategoryRepository(db *sql.DB) *TokenCategoryRepository {
	return &TokenCategoryRepository{db}
}

func (r *TokenCategoryRepository) InsertTokenCategory(tokenCategory models.TokenCategory) (string, error) {
	sqlStatement := `INSERT INTO token_categories (
		title,
		description,
		icon,
		status,
		updated_at,
		created_at
	  ) VALUES (
		$1, $2, $3, $4, $5, $6
	  )
	  RETURNING id`

	var id string

	err := r.db.QueryRow(sqlStatement, tokenCategory.Title, tokenCategory.Description, tokenCategory.Icon, tokenCategory.Status, tokenCategory.UpdatedAt, tokenCategory.CreatedAt).Scan(&id)

	if err != nil {
		return id, err
	}

	return id, nil
}

func (r *TokenCategoryRepository) GetTokenCategoryList(offset int, limit int, keyword string, status *string, orderBy string, orderOption string) ([]models.TokenCategory, error) {
	var tokenCategories []models.TokenCategory

	sqlStatement := `SELECT id, title, description, icon, updated_at, created_at 
					FROM token_categories 
					WHERE title LIKE '%' || $1 || '%' AND (status=$2 OR $2 IS NULL)
					ORDER BY ` + orderBy + ` ` + orderOption + ` 
					OFFSET $3 
					LIMIT $4`

	rows, err := r.db.Query(sqlStatement, keyword, status, offset, limit)

	if err != nil {
		return tokenCategories, err
	}

	defer rows.Close()
	for rows.Next() {
		var tokenCategory models.TokenCategory
		err = rows.Scan(&tokenCategory.ID, &tokenCategory.Title, &tokenCategory.Description, &tokenCategory.Icon, &tokenCategory.UpdatedAt, &tokenCategory.CreatedAt)

		if err != nil {
			return tokenCategories, err
		}

		tokenCategories = append(tokenCategories, tokenCategory)
	}

	return tokenCategories, nil
}

func (r *TokenCategoryRepository) GetTokenCategoryData(id uuid.UUID) (models.TokenCategory, error) {
	sqlStatement := `SELECT id, title, description, icon, updated_at, created_at FROM token_categories WHERE id = $1`

	var tokenCategory models.TokenCategory
	rows, err := r.db.Query(sqlStatement, id)

	if err != nil {
		return tokenCategory, err
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&tokenCategory.ID, &tokenCategory.Title, &tokenCategory.Description, &tokenCategory.Icon, &tokenCategory.UpdatedAt, &tokenCategory.CreatedAt)

		if err != nil {
			return tokenCategory, err
		}
	}

	return tokenCategory, nil
}

func (r *TokenCategoryRepository) UpdateTokenCategory(id uuid.UUID, tokenCategory models.TokenCategory) error {
	sqlStatement := `UPDATE token_categories
	SET title = $2, description = $3, icon = $4, status = $5, updated_at = $6
	WHERE id = $1;`

	_, err := r.db.Exec(sqlStatement, id, tokenCategory.Title, tokenCategory.Description, tokenCategory.Icon, tokenCategory.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *TokenCategoryRepository) DeleteTokenCategory(id uuid.UUID) error {
	sqlStatement := `DELETE FROM token_categories WHERE id = $1`

	_, err := r.db.Exec(sqlStatement, id)

	if err != nil {
		return err
	}

	return nil
}
