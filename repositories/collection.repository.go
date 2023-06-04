package repositories

import (
	"database/sql"
	"metaedu-marketplace/helpers"
	models "metaedu-marketplace/models"

	"github.com/google/uuid"
)

type CollectionRepository struct {
	db *sql.DB
}

func NewCollectionRepository(db *sql.DB) *CollectionRepository {
	return &CollectionRepository{db}
}

func (r *CollectionRepository) InsertCollection(collection models.Collection) (string, error) {
	sqlStatement := `INSERT INTO collections (
		previous_id,
		thumbnail,
		cover,
		title,
		views,
		number_of_items,
		number_of_transactions,
		volume_transactions,
		description,
		category_id,
		creator_id,
		status,
		transaction_hash,
		updated_at,
		created_at
	  ) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
	  )
	  RETURNING id`

	var id string

	err := r.db.QueryRow(sqlStatement, collection.PreviousID, collection.Thumbnail, collection.Cover, collection.Title, collection.Views, collection.NumberOfItems, collection.NumberOfTransactions, collection.VolumeTransactions, collection.Description, collection.CategoryID, collection.CreatorID, collection.Status, collection.UpdatedAt, collection.CreatedAt).Scan(&id)

	if err != nil {
		return id, err
	}

	return id, nil
}

func (r *CollectionRepository) GetCollectionList(offset int, limit int, keyword string, creatorID *uuid.UUID, status *string, orderBy string, orderOption string) ([]models.Collection, error) {
	var collections []models.Collection

	sqlStatement := `SELECT collections.id, collections.previous_id, collections.thumbnail, collections.cover, collections.title, collections.views, collections.number_of_items, collections.number_of_transactions, collections.volume_transactions, collections.description, collections.creator_id, collections.category_id, collections.status, collections.transaction_hash, collections.updated_at, collections.created_at,
						users.id, users.name, users.email, users.photo, users.role, users.address,					
						token_categories.id, token_categories.title, token_categories.description, token_categories.icon, token_categories.updated_at, token_categories.created_at
					FROM collections 
					INNER JOIN users ON users.id = collections.creator_id
					INNER JOIN token_categories ON token_categories.id = collections.category_id
					WHERE collections.title LIKE '%' || $1 || '%' AND (collections.creator_id = $2 OR $2 IS NULL) AND (collections.status = $3 OR $3 IS NULL)
					ORDER BY collections.` + orderBy + ` ` + orderOption + ` 
					OFFSET $4 
					LIMIT $5`

	rows, err := r.db.Query(sqlStatement, keyword, helpers.GetOptionalUUIDParams(creatorID), status, offset, limit)

	if err != nil {
		return collections, err
	}

	defer rows.Close()
	for rows.Next() {
		var collection models.Collection
		err = rows.Scan(&collection.ID, &collection.PreviousID, &collection.Thumbnail, &collection.Cover, &collection.Title, &collection.Views, &collection.NumberOfItems, &collection.NumberOfTransactions, &collection.VolumeTransactions, &collection.Description, &collection.CreatorID, &collection.CategoryID, &collection.Status, &collection.TransactionHash, &collection.UpdatedAt, &collection.CreatedAt,
			&collection.Creator.ID, &collection.Creator.Name, &collection.Creator.Email, &collection.Creator.Photo, &collection.Creator.Role, &collection.Creator.Address,
			&collection.Category.ID, &collection.Category.Title, &collection.Category.Description, &collection.Category.Icon, &collection.Category.UpdatedAt, &collection.Category.CreatedAt)
		if err != nil {
			return collections, err
		}

		collections = append(collections, collection)
	}

	return collections, nil
}

func (r *CollectionRepository) GetCollectionData(id uuid.UUID) (models.Collection, error) {
	sqlStatement := `SELECT collections.id, collections.previous_id, collections.thumbnail, collections.cover, collections.title, collections.views, collections.number_of_items, collections.number_of_transactions, collections.volume_transactions, collections.description, collections.creator_id, collections.category_id, collections.status, collections.transaction_hash, collections.updated_at, collections.created_at,
					users.id, users.name, users.email, users.photo, users.role, users.address	
					FROM collections 
					INNER JOIN users ON collections.creator_id = users.id
					WHERE collections.id = $1`

	var collection models.Collection
	rows, err := r.db.Query(sqlStatement, id)

	if err != nil {
		return collection, err
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&collection.ID, &collection.PreviousID, &collection.Thumbnail, &collection.Cover, &collection.Title, &collection.Views, &collection.NumberOfItems, &collection.NumberOfTransactions, &collection.VolumeTransactions, &collection.Description, &collection.CreatorID, &collection.CategoryID, &collection.Status, &collection.TransactionHash, &collection.UpdatedAt, &collection.CreatedAt,
			&collection.Creator.ID, &collection.Creator.Name, &collection.Creator.Email, &collection.Creator.Photo, &collection.Creator.Role, &collection.Creator.Address)

		if err != nil {
			return collection, err
		}
	}

	return collection, nil
}

func (r *CollectionRepository) UpdateCollection(id uuid.UUID, collection models.Collection) error {
	sqlStatement := `UPDATE collections
	SET thumbnail = $2, cover = $3, title = $4, views = $5, number_of_items = $6, number_of_transactions = $7, volume_transactions = $8, description = $9, category_id = $10, creator_id = $11, status = $12, updated_at = $13
	WHERE id = $1;`

	_, err := r.db.Exec(sqlStatement, id, collection.Thumbnail, collection.Cover, collection.Title, collection.Views, collection.NumberOfItems, collection.NumberOfTransactions, collection.VolumeTransactions, collection.Description, collection.CategoryID, collection.CreatorID, collection.Status, collection.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *CollectionRepository) DeleteCollection(id uuid.UUID) error {
	sqlStatement := `DELETE FROM collections WHERE id = $1`

	_, err := r.db.Exec(sqlStatement, id)

	if err != nil {
		return err
	}

	return nil
}
