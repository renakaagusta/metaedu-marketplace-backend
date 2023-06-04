package repositories

import (
	"database/sql"
	models "metaedu-marketplace/models"

	"github.com/google/uuid"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db}
}

func (r *TransactionRepository) InsertTransaction(transaction models.Transaction) (string, error) {
	sqlStatement := `INSERT INTO transactions (
		previous_id,
		user_from_id,
		user_to_id,
		ownership_id,
		token_id,
		collection_id,
		type,
		quantity,
		amount,
		gas_fee,
		status,
		transaction_hash,
		updated_at,
		created_at
	  ) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
	  )
	  RETURNING id`

	var id string

	err := r.db.QueryRow(sqlStatement,
		transaction.PreviousID,
		transaction.UserFromID,
		transaction.UserToID,
		transaction.OwnershipID,
		transaction.TokenID,
		transaction.CollectionID,
		transaction.Type,
		transaction.Quantity,
		transaction.Amount,
		transaction.GasFee,
		transaction.Status,
		transaction.TransactionHash,
		transaction.UpdatedAt,
		transaction.CreatedAt).Scan(&id)

	if err != nil {
		return id, err
	}

	return id, nil
}

func (r *TransactionRepository) GetTransactionList(offset int, limit int, userID *uuid.UUID, status *string, orderBy string, orderOption string) ([]models.Transaction, error) {
	var transactions []models.Transaction

	sqlStatement := `SELECT transactions.id, transactions.previous_id, transactions.user_from_id, transactions.user_to_id, transactions.ownership_id, transactions.Token_id, transactions.type, transactions.quantity, transactions.amount, transactions.gas_fee, transactions.status, transactions.transaction_hash, transactions.updated_at, transactions.created_at, 
					ownerships.id, ownerships.Token_id, ownerships.user_id, ownerships.quantity, ownerships.sale_price, ownerships.rent_cost, ownerships.available_for_sale, ownerships.available_for_rent, ownerships.updated_at, ownerships.created_at, 
					rentals.id, rentals.user_id, rentals.owner_id, rentals.Token_id, rentals.ownership_id, rentals.updated_at, rentals.created_at,
					user_from.id, user_from.name, user_from.email, user_from.photo, user_from.role, user_from.address,
					user_to.id, user_to.name, user_to.email, user_to.photo, user_to.role, user_to.address,
					tokens.id, tokens.Token_index, tokens.title, tokens.description, tokens.category_id, tokens.collection_id, tokens.image, tokens.uri, tokens.fraction_id, tokens.supply, tokens.last_price, tokens.initial_price, tokens.views, tokens.number_of_transactions, tokens.volume_transactions, tokens.creator_id, tokens.updated_at, tokens.created_at,
					collections.id, collections.thumbnail, collections.cover, collections.title, collections.views, collections.number_of_items, collections.number_of_transactions, collections.volume_transactions, collections.description, collections.creator_id, collections.category_id, collections.updated_at, collections.created_at
					FROM transactions 
					INNER JOIN ownerships ON transactions.ownership_id=ownerships.id 
					LEFT JOIN rentals ON transactions.rental_id=rentals.id 
					LEFT JOIN users user_from ON transactions.user_from_id=user_from.id 
					LEFT JOIN users user_to ON transactions.user_to_id=user_to.id 
					LEFT JOIN collections ON transactions.collection_id=collections.id
					INNER JOIN tokens ON transactions.token_id=tokens.id
					WHERE (transactions.status = $2 OR $2 IS NULL) AND ((transactions.user_from_id=$1 OR $1 IS NULL) OR (transactions.user_to_id=$1 OR $1 IS NULL))
					ORDER BY transactions.` + orderBy + ` ` + orderOption + ` 
					OFFSET $3 
					LIMIT $4`

	rows, err := r.db.Query(sqlStatement, userID, status, offset, limit)

	if err != nil {
		return transactions, err
	}

	defer rows.Close()
	for rows.Next() {
		var transaction models.Transaction
		err = rows.Scan(&transaction.ID, &transaction.PreviousID, &transaction.UserFromID, &transaction.UserToID, &transaction.OwnershipID, &transaction.TokenID, &transaction.Type, &transaction.Quantity, &transaction.Amount, &transaction.GasFee, &transaction.Status, &transaction.TransactionHash, &transaction.UpdatedAt, &transaction.CreatedAt,
			&transaction.Ownership.ID, &transaction.Ownership.TokenID, &transaction.Ownership.UserID, &transaction.Ownership.Quantity, &transaction.Ownership.SalePrice, &transaction.Ownership.RentCost, &transaction.Ownership.AvailableForSale, &transaction.Ownership.AvailableForRent, &transaction.Ownership.UpdatedAt, &transaction.Ownership.CreatedAt,
			&transaction.Rental.ID, &transaction.Rental.UserID, &transaction.Rental.OwnerID, &transaction.Rental.TokenID, &transaction.Rental.OwnershipID, &transaction.Rental.UpdatedAt, &transaction.Rental.CreatedAt,
			&transaction.UserFrom.ID, &transaction.UserFrom.Name, &transaction.UserFrom.Email, &transaction.UserFrom.Photo, &transaction.UserFrom.Role, &transaction.UserFrom.Address,
			&transaction.UserTo.ID, &transaction.UserTo.Name, &transaction.UserTo.Email, &transaction.UserTo.Photo, &transaction.UserTo.Role, &transaction.UserTo.Address,
			&transaction.Token.ID, &transaction.Token.TokenIndex, &transaction.Token.Title, &transaction.Token.Description, &transaction.Token.CategoryID, &transaction.Token.CollectionID, &transaction.Token.Image, &transaction.Token.Uri, &transaction.Token.FractionID, &transaction.Token.Supply, &transaction.Token.LastPrice, &transaction.Token.InitialPrice, &transaction.Token.Views, &transaction.Token.NumberOfTransactions, &transaction.Token.VolumeTransactions, &transaction.Token.CreatorID, &transaction.Token.UpdatedAt, &transaction.Token.CreatedAt,
			&transaction.Collection.ID, &transaction.Collection.Thumbnail, &transaction.Collection.Cover, &transaction.Collection.Title, &transaction.Collection.Views, &transaction.Collection.NumberOfItems, &transaction.Collection.NumberOfTransactions, &transaction.Collection.VolumeTransactions, &transaction.Collection.Description, &transaction.Collection.CreatorID, &transaction.Collection.CategoryID, &transaction.Collection.UpdatedAt, &transaction.Collection.CreatedAt)

		if err != nil {
			return transactions, err
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (r *TransactionRepository) GetTransactionListByToken(offset int, limit int, tokenID uuid.UUID, status *string, orderBy string, orderOption string) ([]models.Transaction, error) {
	var transactions []models.Transaction

	sqlStatement := `SELECT transactions.id, transactions.user_from_id, transactions.user_to_id, transactions.ownership_id, transactions.Token_id, transactions.type, transactions.quantity, transactions.amount, transactions.gas_fee, transactions.transaction_hash, transactions.updated_at, transactions.created_at, 
					ownerships.id, ownerships.Token_id, ownerships.user_id, ownerships.quantity, ownerships.sale_price, ownerships.rent_cost, ownerships.available_for_sale, ownerships.available_for_rent, ownerships.updated_at, ownerships.created_at, 
					rentals.id, rentals.user_id, rentals.owner_id, rentals.Token_id, rentals.ownership_id, rentals.updated_at, rentals.created_at,
					user_from.id, user_from.name, user_from.email, user_from.photo, user_from.role, user_from.address,
					user_to.id, user_to.name, user_to.email, user_to.photo, user_to.role, user_to.address,
					tokens.id, tokens.Token_index, tokens.title, tokens.description, tokens.category_id, tokens.collection_id, tokens.image, tokens.uri, tokens.fraction_id, tokens.supply, tokens.last_price, tokens.initial_price, tokens.views, tokens.number_of_transactions, tokens.volume_transactions, tokens.creator_id, tokens.updated_at, tokens.created_at
					FROM transactions 
					INNER JOIN ownerships ON transactions.ownership_id=ownerships.id 
					LEFT JOIN rentals ON transactions.rental_id=rentals.id 
					LEFT JOIN users user_from ON transactions.user_from_id=user_from.id 
					LEFT JOIN users user_to ON transactions.user_to_id=user_to.id 
					INNER JOIN tokens ON transactions.token_id=tokens.id
					WHERE transactions.token_id = $1 AND (transactions.status=$2 OR $2 IS NULL)
					ORDER BY transactions.` + orderBy + ` ` + orderOption + ` 
					OFFSET $3 
					LIMIT $4`

	rows, err := r.db.Query(sqlStatement, tokenID, status, offset, limit)

	if err != nil {
		return transactions, err
	}

	defer rows.Close()
	for rows.Next() {
		var transaction models.Transaction
		err = rows.Scan(&transaction.ID, &transaction.UserFromID, &transaction.UserToID, &transaction.OwnershipID, &transaction.TokenID, &transaction.Type, &transaction.Quantity, &transaction.Amount, &transaction.GasFee, &transaction.TransactionHash, &transaction.UpdatedAt, &transaction.CreatedAt,
			&transaction.Ownership.ID, &transaction.Ownership.TokenID, &transaction.Ownership.UserID, &transaction.Ownership.Quantity, &transaction.Ownership.SalePrice, &transaction.Ownership.RentCost, &transaction.Ownership.AvailableForSale, &transaction.Ownership.AvailableForRent, &transaction.Ownership.UpdatedAt, &transaction.Ownership.CreatedAt,
			&transaction.Rental.ID, &transaction.Rental.UserID, &transaction.Rental.OwnerID, &transaction.Rental.TokenID, &transaction.Rental.OwnershipID, &transaction.Rental.UpdatedAt, &transaction.Rental.CreatedAt,
			&transaction.UserFrom.ID, &transaction.UserFrom.Name, &transaction.UserFrom.Email, &transaction.UserFrom.Photo, &transaction.UserFrom.Role, &transaction.UserFrom.Address,
			&transaction.UserTo.ID, &transaction.UserTo.Name, &transaction.UserTo.Email, &transaction.UserTo.Photo, &transaction.UserTo.Role, &transaction.UserTo.Address,
			&transaction.Token.ID, &transaction.Token.TokenIndex, &transaction.Token.Title, &transaction.Token.Description, &transaction.Token.CategoryID, &transaction.Token.CollectionID, &transaction.Token.Image, &transaction.Token.Uri, &transaction.Token.FractionID, &transaction.Token.Supply, &transaction.Token.LastPrice, &transaction.Token.InitialPrice, &transaction.Token.Views, &transaction.Token.NumberOfTransactions, &transaction.Token.VolumeTransactions, &transaction.Token.CreatorID, &transaction.Token.UpdatedAt, &transaction.Token.CreatedAt)

		if err != nil {
			return transactions, err
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (r *TransactionRepository) GetTransactionListByCollection(offset int, limit int, collectionID uuid.UUID, status *string, orderBy string, orderOption string) ([]models.Transaction, error) {
	var transactions []models.Transaction

	sqlStatement := `SELECT transactions.id, transactions.user_from_id, transactions.user_to_id, transactions.ownership_id, transactions.Token_id, transactions.type, transactions.quantity, transactions.amount, transactions.gas_fee, transactions.transaction_hash, transactions.updated_at, transactions.created_at, 
					ownerships.id, ownerships.Token_id, ownerships.user_id, ownerships.quantity, ownerships.sale_price, ownerships.rent_cost, ownerships.available_for_sale, ownerships.available_for_rent, ownerships.updated_at, ownerships.created_at, 
					rentals.id, rentals.user_id, rentals.owner_id, rentals.Token_id, rentals.ownership_id, rentals.updated_at, rentals.created_at,
					user_from.id, user_from.name, user_from.email, user_from.photo, user_from.role, user_from.address,
					user_to.id, user_to.name, user_to.email, user_to.photo, user_to.role, user_to.address,
					tokens.id, tokens.Token_index, tokens.title, tokens.description, tokens.category_id, tokens.collection_id, tokens.image, tokens.uri, tokens.fraction_id, tokens.supply, tokens.last_price, tokens.initial_price, tokens.views, tokens.number_of_transactions, tokens.volume_transactions, tokens.creator_id, tokens.updated_at, tokens.created_at
					FROM transactions 
					INNER JOIN ownerships ON transactions.ownership_id=ownerships.id 
					LEFT JOIN rentals ON transactions.rental_id=rentals.id 
					LEFT JOIN users user_from ON transactions.user_from_id=user_from.id 
					LEFT JOIN users user_to ON transactions.user_to_id=user_to.id 
					INNER JOIN tokens ON transactions.token_id=tokens.id
					WHERE transactions.collection_id = $1 AND (transactions.status = $2 OR $2 IS NULL)
					ORDER BY transactions.` + orderBy + ` ` + orderOption + ` 
					OFFSET $3 
					LIMIT $4`

	rows, err := r.db.Query(sqlStatement, collectionID, status, offset, limit)

	if err != nil {
		return transactions, err
	}

	defer rows.Close()
	for rows.Next() {
		var transaction models.Transaction
		err = rows.Scan(&transaction.ID, &transaction.UserFromID, &transaction.UserToID, &transaction.OwnershipID, &transaction.TokenID, &transaction.Type, &transaction.Quantity, &transaction.Amount, &transaction.GasFee, &transaction.TransactionHash, &transaction.UpdatedAt, &transaction.CreatedAt,
			&transaction.Ownership.ID, &transaction.Ownership.TokenID, &transaction.Ownership.UserID, &transaction.Ownership.Quantity, &transaction.Ownership.SalePrice, &transaction.Ownership.RentCost, &transaction.Ownership.AvailableForSale, &transaction.Ownership.AvailableForRent, &transaction.Ownership.UpdatedAt, &transaction.Ownership.CreatedAt,
			&transaction.Rental.ID, &transaction.Rental.UserID, &transaction.Rental.OwnerID, &transaction.Rental.TokenID, &transaction.Rental.OwnershipID, &transaction.Rental.UpdatedAt, &transaction.Rental.CreatedAt,
			&transaction.UserFrom.ID, &transaction.UserFrom.Name, &transaction.UserFrom.Email, &transaction.UserFrom.Photo, &transaction.UserFrom.Role, &transaction.UserFrom.Address,
			&transaction.UserTo.ID, &transaction.UserTo.Name, &transaction.UserTo.Email, &transaction.UserTo.Photo, &transaction.UserTo.Role, &transaction.UserTo.Address,
			&transaction.Token.ID, &transaction.Token.TokenIndex, &transaction.Token.Title, &transaction.Token.Description, &transaction.Token.CategoryID, &transaction.Token.CollectionID, &transaction.Token.Image, &transaction.Token.Uri, &transaction.Token.FractionID, &transaction.Token.Supply, &transaction.Token.LastPrice, &transaction.Token.InitialPrice, &transaction.Token.Views, &transaction.Token.NumberOfTransactions, &transaction.Token.VolumeTransactions, &transaction.Token.CreatorID, &transaction.Token.UpdatedAt, &transaction.Token.CreatedAt)

		if err != nil {
			return transactions, err
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (r *TransactionRepository) GetTransactionData(id uuid.UUID) (models.Transaction, error) {
	sqlStatement := `SELECT id, previous_id, user_from_id, user_to_id, ownership_id, rental_id, token_id, collection_id, type, quantity, amount, gas_fee, status, transaction_hash, updated_at, created_at FROM transactions WHERE id = $1`

	var transaction models.Transaction
	rows, err := r.db.Query(sqlStatement, id)

	if err != nil {
		return transaction, err
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&transaction.ID, &transaction.UserFromID, &transaction.UserToID, &transaction.OwnershipID, &transaction.RentalID, &transaction.TokenID, &transaction.CollectionID, &transaction.Type, &transaction.Quantity, &transaction.Amount, &transaction.GasFee, &transaction.Status, &transaction.TransactionHash, &transaction.UpdatedAt, &transaction.CreatedAt)

		if err != nil {
			return transaction, err
		}
	}

	return transaction, nil
}

func (r *TransactionRepository) UpdateTransaction(id uuid.UUID, transaction models.Transaction) error {
	sqlStatement := `UPDATE transactions
	SET user_from_id = $2, user_to_id = $3, ownership_id = $4, token_id = $5, type = $6, quantity = $7, amount = $8, gas_fee = $9, status = $10, transaction_hash = $11, updated_at = $12
	WHERE id = $1;`

	_, err := r.db.Exec(sqlStatement, id, transaction.UserFromID, transaction.UserToID, transaction.OwnershipID, transaction.TokenID, transaction.Type, transaction.Quantity, transaction.Amount, transaction.GasFee, transaction.Status, transaction.TransactionHash, transaction.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *TransactionRepository) DeleteTransaction(id uuid.UUID) error {
	sqlStatement := `DELETE FROM transactions WHERE id = $1`

	_, err := r.db.Exec(sqlStatement, id)

	if err != nil {
		return err
	}

	return nil
}
