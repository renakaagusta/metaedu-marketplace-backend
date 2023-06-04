package repositories

import (
	"database/sql"
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

func (r *OwnershipRepository) GetOwnershipList(offset int, limit int, status *string, orderBy string, orderOption string) ([]models.Ownership, error) {
	var ownerships []models.Ownership

	sqlStatement := `SELECT ownerships.id, ownerships.previous_id, ownerships.token_id, ownerships.user_id, ownerships.quantity, ownerships.sale_price, ownerships.rent_cost, ownerships.available_for_sale, ownerships.available_for_rent, ownerships.status, ownerships.transaction_hash, ownerships.updated_at, ownerships.created_at,
					tokens.id, tokens.token_index, tokens.token_index,tokens.title, tokens.description, tokens.category_id, tokens.collection_id, tokens.image, tokens.uri, tokens.fraction_id, tokens.supply, tokens.last_price, tokens.initial_price 
					FROM ownerships 
					INNER JOIN tokens ON ownerships.token_id = tokens.id 
					WHERE ownerships.status = $1 OR $1 IS NULL
					ORDER BY tokens.` + orderBy + ` ` + orderOption + ` 
					OFFSET $2
					LIMIT $3`

	rows, err := r.db.Query(sqlStatement, status, offset, limit)

	if err != nil {
		return ownerships, err
	}

	defer rows.Close()
	for rows.Next() {
		var ownership models.Ownership
		err = rows.Scan(&ownership.ID, &ownership.PreviousID, &ownership.TokenID, &ownership.UserID, &ownership.Quantity, &ownership.SalePrice, &ownership.RentCost, &ownership.AvailableForSale, &ownership.AvailableForRent, &ownership.Status, &ownership.TransactionHash, &ownership.UpdatedAt, &ownership.CreatedAt,
			&ownership.Token.ID, &ownership.Token.TokenIndex, &ownership.Token.TokenIndex, &ownership.Token.Title, &ownership.Token.Description, &ownership.Token.CategoryID, &ownership.Token.CollectionID, &ownership.Token.Image, &ownership.Token.Uri, &ownership.Token.FractionID, &ownership.Token.Supply, &ownership.Token.LastPrice, &ownership.Token.InitialPrice)

		if err != nil {
			return ownerships, err
		}

		ownerships = append(ownerships, ownership)
	}

	return ownerships, nil
}

func (r *OwnershipRepository) GetOwnershipListByUserID(offset int, limit int, userID uuid.UUID, status string, orderBy string, orderOption string) ([]models.Ownership, error) {
	var ownerships []models.Ownership

	sqlStatement := `SELECT ownerships.id, ownerships.token_id, ownerships.user_id, ownerships.quantity, ownerships.sale_price, ownerships.rent_cost, ownerships.available_for_sale, ownerships.available_for_rent, ownerships.updated_at, ownerships.created_at, ownerships.status,
					tokens.id, tokens.token_index, tokens.token_index,tokens.title, tokens.description, tokens.category_id, tokens.collection_id, tokens.image, tokens.uri, tokens.fraction_id, tokens.supply, tokens.last_price, tokens.initial_price 
					FROM ownerships 
					INNER JOIN tokens ON ownerships.token_id = tokens.id 
					WHERE ownerships.user_id = $1 AND ownerships.status = $2
					ORDER BY tokens.` + orderBy + ` ` + orderOption + ` 
					OFFSET $3 
					LIMIT $4`

	rows, err := r.db.Query(sqlStatement, userID, status, offset, limit)

	if err != nil {
		return ownerships, err
	}

	defer rows.Close()
	for rows.Next() {
		var ownership models.Ownership
		err = rows.Scan(&ownership.ID, &ownership.TokenID, &ownership.UserID, &ownership.Quantity, &ownership.SalePrice, &ownership.RentCost, &ownership.AvailableForSale, &ownership.AvailableForRent, &ownership.UpdatedAt, &ownership.CreatedAt, &ownership.Status,
			&ownership.Token.ID, &ownership.Token.TokenIndex, &ownership.Token.TokenIndex, &ownership.Token.Title, &ownership.Token.Description, &ownership.Token.CategoryID, &ownership.Token.CollectionID, &ownership.Token.Image, &ownership.Token.Uri, &ownership.Token.FractionID, &ownership.Token.Supply, &ownership.Token.LastPrice, &ownership.Token.InitialPrice)

		if err != nil {
			return ownerships, err
		}

		ownerships = append(ownerships, ownership)
	}

	return ownerships, nil
}

func (r *OwnershipRepository) GetOwnershipListByTokenID(offset int, limit int, tokenID uuid.UUID, status string, orderBy string, orderOption string) ([]models.Ownership, error) {
	var ownerships []models.Ownership

	sqlStatement := `SELECT ownerships.id, ownerships.token_id, ownerships.user_id, ownerships.quantity, ownerships.sale_price, ownerships.rent_cost, ownerships.available_for_sale, ownerships.available_for_rent, ownerships.updated_at, ownerships.created_at, ownerships.status,
				tokens.id, tokens.token_index, tokens.title, tokens.description, tokens.category_id, tokens.collection_id, tokens.image, tokens.uri, tokens.fraction_id, tokens.supply, tokens.last_price, tokens.initial_price, 
				users.id, users.name, users.email, users.photo, users.verified, users.role, users.address from ownerships 
				INNER JOIN tokens ON ownerships.token_id = tokens.id 
				LEFT JOIN users ON ownerships.user_id  = users.id 
				WHERE ownerships.token_id = $1 AND ownerships.status = $2
				ORDER BY tokens.` + orderBy + ` ` + orderOption + ` OFFSET $3 LIMIT $4`

	rows, err := r.db.Query(sqlStatement, tokenID, status, offset, limit)

	if err != nil {
		return ownerships, err
	}

	defer rows.Close()
	for rows.Next() {
		var ownership models.Ownership
		err = rows.Scan(&ownership.ID, &ownership.TokenID, &ownership.UserID, &ownership.Quantity, &ownership.SalePrice, &ownership.RentCost, &ownership.AvailableForSale, &ownership.AvailableForRent, &ownership.UpdatedAt, &ownership.CreatedAt, &ownership.Status,
			&ownership.Token.ID, &ownership.Token.TokenIndex, &ownership.Token.Title, &ownership.Token.Description, &ownership.Token.CategoryID, &ownership.Token.CollectionID, &ownership.Token.Image, &ownership.Token.Uri, &ownership.Token.FractionID, &ownership.Token.Supply, &ownership.Token.LastPrice, &ownership.Token.InitialPrice,
			&ownership.User.ID, &ownership.User.Name, &ownership.User.Email, &ownership.User.Photo, &ownership.User.Verified, &ownership.User.Role, &ownership.User.Address)

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
