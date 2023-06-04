package repositories

import (
	"database/sql"
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

func (r *RentalRepository) GetRentalList(offset int, limit int, status *string, orderBy string, orderOption string) ([]models.Rental, error) {
	var rentals []models.Rental

	sqlStatement := `SELECT rentals.id, rentals.token_id, rentals.ownership_id, rentals.user_id, rentals.owner_id, rentals.timestamp, rentals.status, rentals.transaction_hash, rentals.updated_at, rentals.created_at, 
				tokens.id, tokens.token_index, tokens.title, tokens.description, tokens.category_id, tokens.collection_id, tokens.image, tokens.uri, tokens.fraction_id, tokens.supply, tokens.last_price, tokens.initial_price,
				users.id, users.name, users.email, users.photo, users.verified, users.role, users.address
				FROM rentals
				INNER JOIN tokens ON rentals.token_id = tokens.id 
				LEFT JOIN users ON rentals.user_id = users.id 
				WHERE rentals.status = $1 OR $1 IS NULL
				ORDER BY tokens.` + orderBy + ` ` + orderOption + ` 
				OFFSET $2
				LIMIT $3`
	rows, err := r.db.Query(sqlStatement, status, offset, limit)

	if err != nil {
		return rentals, err
	}

	defer rows.Close()
	for rows.Next() {
		var rental models.Rental
		err = rows.Scan(&rental.ID, &rental.TokenID, &rental.OwnershipID, &rental.UserID, &rental.OwnerID, &rental.Timestamp, &rental.Status, &rental.TransactionHash, &rental.UpdatedAt, &rental.CreatedAt,
			&rental.Token.ID, &rental.Token.TokenIndex, &rental.Token.Title, &rental.Token.Description, &rental.Token.CategoryID, &rental.Token.CollectionID, &rental.Token.Image, &rental.Token.Uri, &rental.Token.FractionID, &rental.Token.Supply, &rental.Token.LastPrice, &rental.Token.InitialPrice,
			&rental.User.ID, &rental.User.Name, &rental.User.Email, &rental.User.Photo, &rental.User.Verified, &rental.User.Role, &rental.User.Address)
		if err != nil {
			return rentals, err
		}

		rentals = append(rentals, rental)
	}

	return rentals, nil
}

func (r *RentalRepository) GetRentalListByUserID(offset int, limit int, userID uuid.UUID, orderBy string, orderOption string, status *string) ([]models.Rental, error) {
	var rentals []models.Rental

	sqlStatement := `SELECT rentals.id, rentals.user_id, rentals.owner_id, rentals.token_id, rentals.ownership_id, rentals.timestamp, rentals.updated_at, rentals.created_at,
	tokens.id, tokens.token_index, tokens.title, tokens.description, tokens.category_id, tokens.collection_id, tokens.image, tokens.uri, tokens.fraction_id, tokens.supply, tokens.last_price, tokens.initial_price,
	borrower.id, borrower.name, borrower.email, borrower.photo, borrower.verified, borrower.role, borrower.address,
	owner.id, owner.name, owner.email, owner.photo, owner.verified, owner.role, owner.address
	FROM rentals
	INNER JOIN tokens ON rentals.token_id=tokens.id
	LEFT JOIN users borrower ON rentals.user_id=borrower.id
	LEFT JOIN users owner ON rentals.owner_id=owner.id
	WHERE rentals.user_id=$1 OR rentals.owner_id=$1 AND (rentals.status = $2 OR $2 IS NULL)
	ORDER BY rentals.` + orderBy + ` ` + orderOption + ` 
	OFFSET $3
	LIMIT $4`

	rows, err := r.db.Query(sqlStatement, userID, status, offset, limit)

	if err != nil {
		return rentals, err
	}

	defer rows.Close()
	for rows.Next() {
		var rental models.Rental
		err = rows.Scan(&rental.ID, &rental.UserID, &rental.OwnerID, &rental.TokenID, &rental.OwnershipID, &rental.Timestamp, &rental.UpdatedAt, &rental.CreatedAt,
			&rental.Token.ID, &rental.Token.TokenIndex, &rental.Token.Title, &rental.Token.Description, &rental.Token.CategoryID, &rental.Token.CollectionID, &rental.Token.Image, &rental.Token.Uri, &rental.Token.FractionID, &rental.Token.Supply, &rental.Token.LastPrice, &rental.Token.InitialPrice,
			&rental.User.ID, &rental.User.Name, &rental.User.Email, &rental.User.Photo, &rental.User.Verified, &rental.User.Role, &rental.User.Address,
			&rental.Owner.ID, &rental.Owner.Name, &rental.Owner.Email, &rental.Owner.Photo, &rental.Owner.Verified, &rental.Owner.Role, &rental.Owner.Address)

		if err != nil {
			return rentals, err
		}

		rentals = append(rentals, rental)
	}

	return rentals, nil
}

func (r *RentalRepository) GetRentalListByTokenID(offset int, limit int, tokenID uuid.UUID, orderBy string, orderOption string, status *string) ([]models.Rental, error) {
	var rentals []models.Rental

	sqlStatement := `SELECT rentals.id, rentals.user_id, rentals.owner_id, rentals.token_id, rentals.ownership_id, rentals.timestamp, rentals.updated_at, rentals.created_at,
	tokens.id, tokens.token_index, tokens.title, tokens.description, tokens.category_id, tokens.collection_id, tokens.image, tokens.uri, tokens.fraction_id, tokens.supply, tokens.last_price, tokens.initial_price,
	borrower.id, borrower.name, borrower.email, borrower.photo, borrower.verified, borrower.role, borrower.address,
	owner.id, owner.name, owner.email, owner.photo, owner.verified, owner.role, owner.address
	FROM rentals
	INNER JOIN tokens ON rentals.token_id=tokens.id
	LEFT JOIN users borrower ON rentals.user_id=borrower.id
	LEFT JOIN users owner ON rentals.owner_id=owner.id
	WHERE rentals.user_id=$1 OR rentals.owner_id=$1 AND (rentals.status = $2 OR $2 IS NULL)
	ORDER BY rentals.` + orderBy + ` ` + orderOption + ` 
	OFFSET $3
	LIMIT $4`

	rows, err := r.db.Query(sqlStatement, tokenID, status, offset, limit)

	if err != nil {
		return rentals, err
	}

	defer rows.Close()
	for rows.Next() {
		var rental models.Rental
		err = rows.Scan(&rental.ID, &rental.UserID, &rental.OwnerID, &rental.TokenID, &rental.OwnershipID, &rental.Timestamp, &rental.UpdatedAt, &rental.CreatedAt,
			&rental.Token.ID, &rental.Token.TokenIndex, &rental.Token.Title, &rental.Token.Description, &rental.Token.CategoryID, &rental.Token.CollectionID, &rental.Token.Image, &rental.Token.Uri, &rental.Token.FractionID, &rental.Token.Supply, &rental.Token.LastPrice, &rental.Token.InitialPrice,
			&rental.User.ID, &rental.User.Name, &rental.User.Email, &rental.User.Photo, &rental.User.Verified, &rental.User.Role, &rental.User.Address,
			&rental.Owner.ID, &rental.Owner.Name, &rental.Owner.Email, &rental.Owner.Photo, &rental.Owner.Verified, &rental.Owner.Role, &rental.Owner.Address)

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
