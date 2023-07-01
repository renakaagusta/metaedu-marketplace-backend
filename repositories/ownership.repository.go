package repositories

import (
	"database/sql"
	"metaedu-marketplace/helpers"
	models "metaedu-marketplace/models"

	"github.com/google/uuid"
)

type OwnershipRepository struct {
	db *sql.DB
}

func NewOwnershipRepository(db *sql.DB) *OwnershipRepository {
	return &OwnershipRepository{db}
}

func (r *OwnershipRepository) InsertOwnership(ownership models.Ownership) (string, error) {
	sqlStatement := `INSERT INTO ownerships (
		previous_id,
		token_id,
		user_id,
		quantity,
		sale_price,
		rent_cost,
		available_for_sale,
		available_for_rent,
		status,
		transaction_hash,
		updated_at,
		created_at
	  ) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
	  )
	  RETURNING id`

	var id string

	err := r.db.QueryRow(sqlStatement, ownership.PreviousID, ownership.TokenID, ownership.UserID, ownership.Quantity, ownership.SalePrice, ownership.RentCost, ownership.AvailableForSale, ownership.AvailableForRent, ownership.Status, ownership.TransactionHash, ownership.UpdatedAt, ownership.CreatedAt).Scan(&id)

	if err != nil {
		return id, err
	}

	return id, nil
}

func (r *OwnershipRepository) GetOwnershipList(offset int, limit int, keyword string, userID *uuid.UUID, user *string, creatorID *uuid.UUID, creator *string, tokenID *uuid.UUID, status string, orderBy string, orderOption string) ([]models.Ownership, error) {
	var ownerships []models.Ownership

	sqlStatement := `SELECT ownerships.id, ownerships.previous_id, ownerships.token_id, ownerships.user_id, ownerships.quantity, ownerships.sale_price, ownerships.rent_cost, ownerships.available_for_sale, ownerships.available_for_rent, ownerships.updated_at, ownerships.created_at, ownerships.status, ownerships.transaction_hash,
				tokens.id, tokens.token_index, tokens.title, tokens.description, tokens.category_id, tokens.collection_id, tokens.image, tokens.uri, tokens.fraction_id, tokens.supply, tokens.last_price, tokens.initial_price, 
				owners.id, owners.name, owners.email, owners.photo, owners.verified, owners.role, owners.address,
				creators.id, creators.name, creators.email, creators.photo, creators.verified, creators.role, creators.address
				FROM ownerships 
				INNER JOIN tokens ON ownerships.token_id=tokens.id 
				LEFT JOIN users owners ON ownerships.user_id=owners.id 
				LEFT JOIN users creators ON tokens.creator_id=creators.id 
				WHERE LOWER(tokens.title) LIKE '%' || LOWER($1) || '%' AND (owners.id = $2 OR $2 IS NULL) AND (owners.address = $3 OR $3 IS NULL) AND (creators.id = $4 OR $4 IS NULL) AND (creators.address = $5 OR $5 IS NULL) AND (tokens.id = $6 OR $6 IS NULL) AND (ownerships.status = $7 OR $7 IS NULL)
				ORDER BY tokens.` + orderBy + ` ` + orderOption + ` 
				OFFSET $8 
				LIMIT $9`

	rows, err := r.db.Query(sqlStatement, keyword, helpers.GetOptionalUUIDParams(userID), helpers.GetOptionalStringParams(user), helpers.GetOptionalUUIDParams(creatorID), helpers.GetOptionalStringParams(creator), helpers.GetOptionalUUIDParams(tokenID), status, offset, limit)

	if err != nil {
		return ownerships, err
	}

	defer rows.Close()
	for rows.Next() {
		var ownership models.Ownership
		err = rows.Scan(&ownership.ID, &ownership.PreviousID, &ownership.TokenID, &ownership.UserID, &ownership.Quantity, &ownership.SalePrice, &ownership.RentCost, &ownership.AvailableForSale, &ownership.AvailableForRent, &ownership.UpdatedAt, &ownership.CreatedAt, &ownership.Status, &ownership.TransactionHash,
			&ownership.Token.ID, &ownership.Token.TokenIndex, &ownership.Token.Title, &ownership.Token.Description, &ownership.Token.CategoryID, &ownership.Token.CollectionID, &ownership.Token.Image, &ownership.Token.Uri, &ownership.Token.FractionID, &ownership.Token.Supply, &ownership.Token.LastPrice, &ownership.Token.InitialPrice,
			&ownership.User.ID, &ownership.User.Name, &ownership.User.Email, &ownership.User.Photo, &ownership.User.Verified, &ownership.User.Role, &ownership.User.Address,
			&ownership.Token.Creator.ID, &ownership.Token.Creator.Name, &ownership.Token.Creator.Email, &ownership.Token.Creator.Photo, &ownership.Token.Creator.Verified, &ownership.Token.Creator.Role, &ownership.Token.Creator.Address)

		if err != nil {
			return ownerships, err
		}

		ownerships = append(ownerships, ownership)
	}

	return ownerships, nil
}

func (r *OwnershipRepository) GetOwnershipByTokenAndUser(tokenID uuid.UUID, userID uuid.UUID) (models.Ownership, error) {
	sqlStatement := `SELECT id, previous_id, token_id, user_id, quantity, sale_price, rent_cost, available_for_sale, available_for_rent, status, transaction_hash, updated_at, created_at FROM ownerships WHERE token_id = $1 AND user_id = $2`

	var ownership models.Ownership
	rows, err := r.db.Query(sqlStatement, tokenID, userID)

	if err != nil {
		return ownership, err
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&ownership.ID, &ownership.PreviousID, &ownership.TokenID, &ownership.UserID, &ownership.Quantity, &ownership.SalePrice, &ownership.RentCost, &ownership.AvailableForSale, &ownership.AvailableForRent, &ownership.Status, &ownership.TransactionHash, &ownership.UpdatedAt, &ownership.CreatedAt)

		if err != nil {
			return ownership, err
		}
	}

	return ownership, nil
}

func (r *OwnershipRepository) GetOwnershipData(id uuid.UUID) (models.Ownership, error) {
	sqlStatement := `SELECT id, previous_id, token_id, user_id, quantity, sale_price, rent_cost, available_for_sale, available_for_rent, status, transaction_hash, updated_at, created_at FROM ownerships WHERE id = $1`

	var ownership models.Ownership
	rows, err := r.db.Query(sqlStatement, id)

	if err != nil {
		return ownership, err
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&ownership.ID, &ownership.PreviousID, &ownership.TokenID, &ownership.UserID, &ownership.Quantity, &ownership.SalePrice, &ownership.RentCost, &ownership.AvailableForSale, &ownership.AvailableForRent, &ownership.Status, &ownership.TransactionHash, &ownership.UpdatedAt, &ownership.CreatedAt)

		if err != nil {
			return ownership, err
		}
	}

	return ownership, nil
}

func (r *OwnershipRepository) UpdateOwnership(id uuid.UUID, ownership models.Ownership) error {
	sqlStatement := `UPDATE ownerships
	SET token_id = $2, user_id = $3, quantity = $4, sale_price = $5, rent_cost = $6, available_for_sale = $7, available_for_rent = $8, status = $9, transaction_hash = $10, updated_at = $11
	WHERE id = $1;`

	_, err := r.db.Exec(sqlStatement, id, ownership.TokenID, ownership.UserID, ownership.Quantity, ownership.SalePrice, ownership.RentCost, ownership.AvailableForSale, ownership.AvailableForRent, ownership.Status, ownership.TransactionHash, ownership.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *OwnershipRepository) DeleteOwnership(id uuid.UUID) error {
	sqlStatement := `DELETE FROM ownerships WHERE id = $1`

	_, err := r.db.Exec(sqlStatement, id)

	if err != nil {
		return err
	}

	return nil
}
