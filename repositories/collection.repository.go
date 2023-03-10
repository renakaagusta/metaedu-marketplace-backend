package repositories

import (
	"database/sql"
	"log"
	models "metaedu-marketplace/models"
)

type CollectionRepository struct {
	db *sql.DB
}

func NewCollectionRepository(db *sql.DB) *CollectionRepository {
	return &CollectionRepository{db}
}

func (r *CollectionRepository) InsertCollection(collection models.Collection) string {
	sqlStatement := `INSERT INTO collections (
		title,
		description,
		owner,
		updated_at
	  ) VALUES (
		$1, $2, $3, $4
	  )
	  RETURNING id`

	var id string

	err := r.db.QueryRow(sqlStatement, collection.Title, collection.Description, collection.Owner, collection.UpdatedAt).Scan(&id)

	if err != nil {
		log.Fatalf("Query is cannot executed %v", err)
	}

	return id
}

func (r *CollectionRepository) UpdateCollection(collection models.Collection) error {
	sqlStatement := `UPDATE collections
	SET title = $2, description = $3, owner = $4, updated_at = $4
	WHERE id = $1;`

	_, err := r.db.Exec(sqlStatement, collection.ID, collection.Title, collection.Description, collection.Owner, collection.UpdatedAt)

	if err != nil {
		log.Fatalf("Query is cannot executed %v", err)
	}

	return err
}
