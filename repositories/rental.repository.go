package repositories

import (
	"database/sql"
	"metaedu-marketplace/helpers"
	models "metaedu-marketplace/models"

	"github.com/google/uuid"
)

type RentalRepository struct {
	db *sql.DB
}

func NewRentalRepository(db *sql.DB) *RentalRepository {
	return &RentalRepository{db}
}

func (r *RentalRepository) InsertRental(rental models.Rental) (string, error) {
	sqlStatement := `INSERT INTO rentals (
		previous_id,
		user_id,
		owner_id,
		token_id,
		ownership_id,
		timestamp,
		status,
		transaction_hash,
		updated_at,
		created_at
	  ) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
	  )
	  RETURNING id`

	var id string

	err := r.db.QueryRow(sqlStatement, rental.PreviousID, rental.UserID, rental.OwnerID, rental.TokenID, rental.OwnershipID, rental.Timestamp, rental.Status, rental.TransactionHash, rental.UpdatedAt, rental.CreatedAt).Scan(&id)

	if err != nil {
		return id, err
	}

	return id, nil
}

func (r *RentalRepository) GetRentalList(offset int, limit int, keyword string, userID *uuid.UUID, user *string, ownerID *uuid.UUID, owner *string, creatorID *uuid.UUID, creator *string, tokenID *uuid.UUID, status *string, orderBy string, orderOption string) ([]models.Rental, error) {
	var rentals []models.Rental

	sqlStatement := `SELECT rentals.id, rentals.previous_id, rentals.token_id, rentals.ownership_id, rentals.user_id, rentals.owner_id, rentals.timestamp, rentals.status, rentals.transaction_hash, rentals.updated_at, rentals.created_at, 
				tokens.id, tokens.token_index, tokens.title, tokens.description, tokens.category_id, tokens.collection_id, tokens.image, tokens.uri, tokens.fraction_id, tokens.supply, tokens.last_price, tokens.initial_price,
				users.id, users.name, users.email, users.photo, users.verified, users.role, users.address,
				owners.id, owners.name, owners.email, owners.photo, owners.verified, owners.role, owners.address,
				creators.id, creators.name, creators.email, creators.photo, creators.verified, creators.role, creators.address
				FROM rentals
				INNER JOIN tokens ON rentals.token_id = tokens.id 
				LEFT JOIN users owners ON rentals.owner_id=owners.id 
				LEFT JOIN users users ON rentals.user_id=users.id 
				LEFT JOIN users creators ON tokens.creator_id=creators.id 
				WHERE LOWER(tokens.title) LIKE '%' || LOWER($1) || '%' AND (users.id = $2 OR $2 IS NULL) AND (users.address = $3 OR $3 IS NULL) AND (owners.id = $4 OR $4 IS NULL) AND (owners.address = $5 OR $5 IS NULL) AND (creators.id = $6 OR $6 IS NULL) AND (creators.address = $7 OR $7 IS NULL) AND (tokens.id = $8 OR $8 IS NULL) AND (rentals.status = $9 OR $9 IS NULL)
				ORDER BY tokens.` + orderBy + ` ` + orderOption + ` 
				OFFSET $10
				LIMIT $11`
	rows, err := r.db.Query(sqlStatement, keyword, helpers.GetOptionalUUIDParams(userID), helpers.GetOptionalStringParams(user), helpers.GetOptionalUUIDParams(ownerID), helpers.GetOptionalStringParams(owner), helpers.GetOptionalUUIDParams(creatorID), helpers.GetOptionalStringParams(creator), helpers.GetOptionalUUIDParams(tokenID), status, offset, limit)

	if err != nil {
		return rentals, err
	}

	defer rows.Close()
	for rows.Next() {
		var rental models.Rental
		err = rows.Scan(&rental.ID, &rental.PreviousID, &rental.TokenID, &rental.OwnershipID, &rental.UserID, &rental.OwnerID, &rental.Timestamp, &rental.Status, &rental.TransactionHash, &rental.UpdatedAt, &rental.CreatedAt,
			&rental.Token.ID, &rental.Token.TokenIndex, &rental.Token.Title, &rental.Token.Description, &rental.Token.CategoryID, &rental.Token.CollectionID, &rental.Token.Image, &rental.Token.Uri, &rental.Token.FractionID, &rental.Token.Supply, &rental.Token.LastPrice, &rental.Token.InitialPrice,
			&rental.User.ID, &rental.User.Name, &rental.User.Email, &rental.User.Photo, &rental.User.Verified, &rental.User.Role, &rental.User.Address,
			&rental.Owner.ID, &rental.Owner.Name, &rental.Owner.Email, &rental.Owner.Photo, &rental.Owner.Verified, &rental.Owner.Role, &rental.Owner.Address,
			&rental.Token.Creator.ID, &rental.Token.Creator.Name, &rental.Token.Creator.Email, &rental.Token.Creator.Photo, &rental.Token.Creator.Verified, &rental.Token.Creator.Role, &rental.Token.Creator.Address)
		if err != nil {
			return rentals, err
		}

		rentals = append(rentals, rental)
	}

	return rentals, nil
}

func (r *RentalRepository) GetRentalData(id uuid.UUID) (models.Rental, error) {
	sqlStatement := `SELECT id, previous_id, user_id, owner_id, token_id, ownership_id, timestamp, status, transaction_hash, updated_at, created_at FROM rentals WHERE id = $1`

	var rental models.Rental
	rows, err := r.db.Query(sqlStatement, id)

	if err != nil {
		return rental, err
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&rental.ID, &rental.PreviousID, &rental.UserID, &rental.OwnerID, &rental.TokenID, &rental.OwnershipID, &rental.Timestamp, &rental.Status, &rental.TransactionHash, &rental.UpdatedAt, &rental.CreatedAt)

		if err != nil {
			return rental, err
		}
	}

	return rental, nil
}

func (r *RentalRepository) UpdateRental(id uuid.UUID, rental models.Rental) error {
	sqlStatement := `UPDATE rentals
	SET user_id = $2, owner_id = $3, token_id = $4, ownership_id = $5, timestamp = $6, status = $7, updated_at = $8
	WHERE id = $1;`

	_, err := r.db.Exec(sqlStatement, id, rental.UserID, rental.OwnerID, rental.TokenID, rental.OwnershipID, rental.Timestamp, rental.Status, rental.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *RentalRepository) DeleteRental(id uuid.UUID) error {
	sqlStatement := `DELETE FROM rentals WHERE id = $1`

	_, err := r.db.Exec(sqlStatement, id)

	if err != nil {
		return err
	}

	return nil
}
