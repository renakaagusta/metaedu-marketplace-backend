package repositories

import (
	"database/sql"
	models "metaedu-marketplace/models"

	"github.com/google/uuid"
)

type FractionRepository struct {
	db *sql.DB
}

func NewFractionRepository(db *sql.DB) *FractionRepository {
	return &FractionRepository{db}
}

func (r *FractionRepository) InsertFraction(fraction models.Fraction) (string, error) {
	sqlStatement := `INSERT INTO fractions (
		previous_id,
		token_parent_id,
		token_fraction_id,
		status,
		transaction_hash,
		updated_at,
		created_at
	  ) VALUES (
		$1, $2, $3, $4, $5, $6, $7
	  )
	  RETURNING id`

	var id string

	err := r.db.QueryRow(sqlStatement, fraction.PreviousID, fraction.TokenParentID, fraction.TokenFractionID, fraction.Status, fraction.TransactionHash, fraction.UpdatedAt, fraction.CreatedAt).Scan(&id)

	if err != nil {
		return id, err
	}

	return id, nil
}

func (r *FractionRepository) GetFractionList(offset int, limit int, status *string, orderBy string, orderOption string) ([]models.Fraction, error) {
	var fractions []models.Fraction

	sqlStatement := `SELECT id, previous_id, token_parent_id, token_fraction_id, status, transaction_hash, updated_at, created_at 
					FROM fractions 
					WHERE (status=$1 OR $1 IS NULL) 
					ORDER BY ` + orderBy + ` ` + orderOption + ` 
					OFFSET $2 
					LIMIT $3`

	rows, err := r.db.Query(sqlStatement, status, offset, limit)

	if err != nil {
		return fractions, err
	}

	defer rows.Close()
	for rows.Next() {
		var fraction models.Fraction
		err = rows.Scan(&fraction.ID, &fraction.PreviousID, &fraction.TokenParentID, &fraction.TokenFractionID, &fraction.Status, &fraction.TransactionHash, &fraction.UpdatedAt, &fraction.CreatedAt)

		if err != nil {
			return fractions, err
		}

		fractions = append(fractions, fraction)
	}

	return fractions, nil
}

func (r *FractionRepository) GetFractionData(id uuid.UUID) (models.Fraction, error) {
	sqlStatement := `SELECT id, previous_id, token_parent_id, token_fraction_id, status, transaction_hash, updated_at, created_at FROM fractions WHERE id = $1`

	var fraction models.Fraction
	rows, err := r.db.Query(sqlStatement, id)

	if err != nil {
		return fraction, err
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&fraction.ID, &fraction.TokenParentID, &fraction.TokenFractionID, &fraction.Status, &fraction.TransactionHash, &fraction.UpdatedAt, &fraction.CreatedAt)

		if err != nil {
			return fraction, err
		}
	}

	return fraction, nil
}

func (r *FractionRepository) UpdateFraction(id uuid.UUID, fraction models.Fraction) error {
	sqlStatement := `UPDATE fractions
	SET token_parent_id = $2, token_fraction_id = $3, status = $4, updated_at = $5
	WHERE id = $1;`

	_, err := r.db.Exec(sqlStatement, id, fraction.TokenParentID, fraction.TokenFractionID, fraction.Status, fraction.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *FractionRepository) DeleteFraction(id uuid.UUID) error {
	sqlStatement := `DELETE FROM fractions WHERE id = $1`

	_, err := r.db.Exec(sqlStatement, id)

	if err != nil {
		return err
	}

	return nil
}
