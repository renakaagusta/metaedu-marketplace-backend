package repositories

import (
	"database/sql"
	helpers "metaedu-marketplace/helpers"
	models "metaedu-marketplace/models"

	"github.com/google/uuid"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db}
}

func (r *TokenRepository) InsertToken(token models.Token, attributesBytes []byte) (models.Token, error) {
	sqlStatement := `INSERT INTO tokens (
		previous_id,
		token_index,
		title,
		description,
		category_id,
		collection_id,
		image,
		uri,
		source_id,
		fraction_id,
		supply,
		last_price,
		initial_price,
		views,
		number_of_transactions,
		volume_transactions,
		creator_id,
		attributes,
		status,
		transaction_hash,
		updated_at,
		created_at
	  ) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
	  )
	  RETURNING id, uri, supply, token_index`

	err := r.db.QueryRow(sqlStatement, token.PreviousID, token.TokenIndex, token.Title, token.Description, token.CategoryID, token.CollectionID, token.Image, token.Uri, token.SourceID, token.FractionID, token.Supply, token.LastPrice, token.InitialPrice, token.Views, token.NumberOfTransactions, token.VolumeTransactions, token.CreatorID, token.Attributes, token.Status, token.TransactionHash, token.UpdatedAt, token.CreatedAt).Scan(&token.ID, &token.Uri, &token.Supply, &token.TokenIndex)

	if err != nil {
		return token, err
	}

	return token, nil
}

func (r *TokenRepository) GetTokenList(offset int, limit int, keyword string, category *uuid.UUID, collection *uuid.UUID, creatorID *uuid.UUID, creator *string, minPrice int, maxPrice int, status *string, orderBy string, orderOption string) ([]models.Token, error) {
	var tokens []models.Token

	sqlStatement := `SELECT tokens.id, tokens.previous_id, tokens.token_index, tokens.title, tokens.description, tokens.category_id, tokens.collection_id, tokens.image, tokens.uri, tokens.source_id, tokens.fraction_id, tokens.supply, tokens.last_price, tokens.initial_price, tokens.views, tokens.number_of_transactions, tokens.volume_transactions, tokens.creator_id, tokens.attributes, tokens.status, tokens.transaction_hash, tokens.updated_at, tokens.created_at,
					users.id, users.name, users.email, users.photo, users.role, users.address	
					FROM tokens 
					INNER JOIN users ON tokens.creator_id=users.id
					WHERE LOWER(tokens.title) LIKE '%' || LOWER($1) || '%' AND (tokens.category_id = $2 OR $2 IS NULL) AND (tokens.collection_id = $3 OR $3 IS NULL) AND (tokens.creator_id = $4 OR $4 IS NULL) AND (users.address = $5 OR $5 IS NULL) AND tokens.status=$6 AND tokens.last_price >= $7 AND tokens.last_price <= $8
					ORDER BY tokens.` + orderBy + ` ` + orderOption + `
					OFFSET $9
					LIMIT $10`

	var rows *sql.Rows
	var err error

	rows, err = r.db.Query(sqlStatement, keyword, helpers.GetOptionalUUIDParams(category), helpers.GetOptionalUUIDParams(collection), helpers.GetOptionalUUIDParams(creatorID), helpers.GetOptionalStringParams(creator), status, minPrice, maxPrice, offset, limit)

	if err != nil {
		return tokens, err
	}

	defer rows.Close()
	for rows.Next() {
		var token models.Token
		err = rows.Scan(&token.ID, &token.PreviousID, &token.TokenIndex, &token.Title, &token.Description, &token.CategoryID, &token.CollectionID, &token.Image, &token.Uri, &token.SourceID, &token.FractionID, &token.Supply, &token.LastPrice, &token.InitialPrice, &token.Views, &token.NumberOfTransactions, &token.VolumeTransactions, &token.CreatorID, &token.Attributes, &token.Status, &token.TransactionHash, &token.UpdatedAt, &token.CreatedAt,
			&token.Creator.ID, &token.Creator.Name, &token.Creator.Email, &token.Creator.Photo, &token.Creator.Role, &token.Creator.Address)

		if err != nil {
			return tokens, err
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (r *TokenRepository) GetLastTokenIndex() (int, error) {
	var tokenIndex int

	sqlStatement := `SELECT token_index 
					FROM tokens 
					WHERE status='active'
					ORDER BY token_index 
					DESC LIMIT 1`

	rows, err := r.db.Query(sqlStatement)

	if err != nil {
		return tokenIndex, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&tokenIndex)

		if err != nil {
			return tokenIndex, err
		}
	}

	return tokenIndex, nil
}

func (r *TokenRepository) GetTokenData(id uuid.UUID) (models.Token, error) {
	sqlStatement := `SELECT tokens.id, tokens.previous_id, tokens.token_index, tokens.title, tokens.description, tokens.category_id, tokens.collection_id, tokens.fraction_id, tokens.source_id, tokens.image, tokens.uri, tokens.source_id, tokens.fraction_id, tokens.supply, tokens.last_price, tokens.initial_price, tokens.views, tokens.number_of_transactions, tokens.volume_transactions, tokens.creator_id, tokens.attributes, tokens.status, tokens.transaction_hash, tokens.updated_at, tokens.created_at,
					users.id, users.name, users.email, users.photo, users.role, users.address,
					token_categories.id, token_categories.title, token_categories.description, token_categories.icon, token_categories.updated_at, token_categories.created_at,
					collections.id, collections.thumbnail, collections.cover, collections.title, collections.views, collections.number_of_items, collections.number_of_transactions, collections.volume_transactions, collections.description, collections.creator_id, collections.category_id, collections.updated_at, collections.created_at
					FROM tokens 
					INNER JOIN token_categories ON tokens.category_id = token_categories.id 
					INNER JOIN users ON tokens.creator_id = users.id
					LEFT JOIN collections ON tokens.collection_id = collections.id 
					WHERE tokens.id = $1`

	var token models.Token
	rows, err := r.db.Query(sqlStatement, id)

	if err != nil {
		return token, err
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&token.ID, &token.PreviousID, &token.TokenIndex, &token.Title, &token.Description, &token.CategoryID, &token.CollectionID, &token.FractionID, &token.SourceID, &token.Image, &token.Uri, &token.SourceID, &token.FractionID, &token.Supply, &token.LastPrice, &token.InitialPrice, &token.Views, &token.NumberOfTransactions, &token.VolumeTransactions, &token.CreatorID, &token.Attributes, &token.Status, &token.TransactionHash, &token.UpdatedAt, &token.CreatedAt,
			&token.Creator.ID, &token.Creator.Name, &token.Creator.Email, &token.Creator.Photo, &token.Creator.Role, &token.Creator.Address,
			&token.Category.ID, &token.Category.Title, &token.Category.Description, &token.Category.Icon, &token.Category.UpdatedAt, &token.Category.CreatedAt, &token.Collection.ID, &token.Collection.Thumbnail, &token.Collection.Cover, &token.Collection.Title, &token.Collection.Views, &token.Collection.NumberOfItems, &token.Collection.NumberOfTransactions, &token.Collection.VolumeTransactions, &token.Collection.Description, &token.Collection.CreatorID, &token.Collection.CategoryID, &token.Collection.UpdatedAt, &token.Collection.CreatedAt)

		if err != nil {
			return token, err
		}
	}

	return token, nil
}

func (r *TokenRepository) UpdateToken(id uuid.UUID, token models.Token) error {
	sqlStatement := `UPDATE tokens
	SET title = $2, description = $3, category_id = $4, collection_id = $5, image = $6, uri = $7, source_id = $8, fraction_id = $9, supply = $10, last_price = $11, initial_price = $12, views = $13, number_of_transactions = $14, volume_transactions = $15, creator_id = $16, status = $17, transaction_hash = $18, updated_at = $19
	WHERE id = $1;`

	_, err := r.db.Exec(sqlStatement, id, token.Title, token.Description, token.CategoryID, token.CollectionID, token.Image, token.Uri, token.SourceID, token.FractionID, token.Supply, token.LastPrice, token.InitialPrice, token.Views, token.NumberOfTransactions, token.VolumeTransactions, token.CreatorID, token.Status, token.TransactionHash, token.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *TokenRepository) DeleteToken(id uuid.UUID) error {
	sqlStatement := `DELETE FROM tokens WHERE id = $1`

	_, err := r.db.Exec(sqlStatement, id)

	if err != nil {
		return err
	}

	return nil
}
