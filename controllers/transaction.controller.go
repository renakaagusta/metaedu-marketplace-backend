package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"metaedu-marketplace/helpers"
	models "metaedu-marketplace/models"
	"metaedu-marketplace/repositories"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/web3-storage/go-w3s-client"
)

type TransactionController struct {
	repository           *repositories.TransactionRepository
	tokenRepository      *repositories.TokenRepository
	collectionRepository *repositories.CollectionRepository
	ownershipRepository  *repositories.OwnershipRepository
	rentalRepository     *repositories.RentalRepository
	web3StorageClient    w3s.Client
	redisClient          *redis.Client
}

func NewTransactionController(repository *repositories.TransactionRepository, tokenRepository *repositories.TokenRepository, collectionRepository *repositories.CollectionRepository, ownershipRepository *repositories.OwnershipRepository, rentalRepository *repositories.RentalRepository, web3StorageClient w3s.Client, redisClient *redis.Client) *TransactionController {
	return &TransactionController{repository, tokenRepository, collectionRepository, ownershipRepository, rentalRepository, web3StorageClient, redisClient}
}

func (ac *TransactionController) InsertTransaction(ctx *gin.Context) {
	// Validate transaction hash
	transactionHash := ctx.DefaultPostForm("transaction_hash", "")

	if transactionHash == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Transaction hash is required"})
		return
	}

	// Validate quantity
	quantity, err := strconv.Atoi(ctx.DefaultPostForm("quantity", "0"))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if quantity < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Quantity must be greater than 1 or equal"})
		return
	}

	// Validate amount
	amount, err := strconv.ParseFloat(ctx.DefaultPostForm("amount", "0"), 64)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if amount <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Amount must be greater than 0"})
		return
	}

	// Validate gas fee
	gasFee, err := strconv.ParseFloat(ctx.DefaultPostForm("gas_fee", "0"), 64)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if gasFee <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Gas fee must be greater than 0"})
		return
	}

	// Validate user from id
	userFromIDParams := ctx.PostForm("user_from_id")
	var userFromID uuid.UUID

	if userFromIDParams != "" {
		userFromID, err = uuid.Parse(userFromIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Category id is not valid"})
			return
		}
	}

	// Validate user to id
	userToIDParams := ctx.PostForm("user_to_id")
	var userToID uuid.UUID

	if userToIDParams != "" {
		userToID, err = uuid.Parse(userToIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Category id is not valid"})
			return
		}
	}

	// Validate token
	tokenIDParams := ctx.PostForm("token_id")
	var tokenID uuid.UUID

	if tokenIDParams != "" {
		tokenID, err = uuid.Parse(tokenIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Token id is not valid"})
			return
		}
	}

	// Validate ownership
	ownershipIDParams := ctx.PostForm("ownership_id")
	var ownershipID uuid.UUID

	if ownershipIDParams != "" {
		ownershipID, err = uuid.Parse(ownershipIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Ownership id is not valid"})
			return
		}
	}

	// Validate
	user, isExist := ctx.Get("user")

	if !isExist {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": "User data is not valid"})
		return
	}

	// Validate type
	transactionType := ctx.PostForm("type")

	if transactionType == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Type is required"})
		return
	}

	status := "active"

	// Check if token is still in rental period
	rentals, err := ac.rentalRepository.GetRentalList(100, 0, "", nil, nil, nil, nil, nil, nil, &tokenID, &status, "created_at", "DESC")

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	for _, rental := range rentals {
		if rental.Timestamp.Time.After(time.Now()) {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Item still in rental period"})
			return
		}
	}

	var transaction models.Transaction

	if transactionType == "purchase" {
		ownership, err := ac.ownershipRepository.GetOwnershipData(ownershipID)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err})
			return
		}

		if !ownership.AvailableForSale {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Token is not available for sale"})
			return
		}

		if quantity > ownership.Quantity {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Quantity is not valid"})
			return
		}

		if amount < ownership.SalePrice*float64(quantity) {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Amount is lower than sale price"})
			return
		}

		ownership.PreviousID = ownership.ID
		ownership.Quantity = ownership.Quantity - quantity
		ownership.Status = "waiting_confirmation"
		ownership.TransactionHash = transactionHash

		_, err = ac.ownershipRepository.InsertOwnership(ownership)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err})
			return
		}

		var newOwnership models.Ownership
		newOwnership.PreviousID = uuid.UUID{}
		newOwnership.UserID = user.(models.User).ID
		newOwnership.TokenID = tokenID
		newOwnership.Quantity = quantity
		newOwnership.AvailableForSale = false
		newOwnership.AvailableForRent = false
		newOwnership.Status = "waiting_confirmation"
		newOwnership.TransactionHash = transactionHash
		newOwnership.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
		newOwnership.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

		newOwnershipIDResult, err := ac.ownershipRepository.InsertOwnership(newOwnership)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err})
			return
		}

		newOwnershipID, err := uuid.Parse(newOwnershipIDResult)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		transaction.OwnershipID = newOwnershipID
	} else if transactionType == "rent" {
		if quantity > 1 {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Maximum quantity for rental is 1"})
			return
		}

		// Validate days
		daysParams := ctx.PostForm("days")

		if daysParams == "" {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Timestamp is required"})
			return
		}

		days, err := strconv.Atoi(daysParams)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
			return
		}

		ownership, err := ac.ownershipRepository.GetOwnershipData(ownershipID)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err})
			return
		}

		if quantity > ownership.Quantity {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Quantity is not valid"})
			return
		}

		if !ownership.AvailableForRent {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Token is not available for rent"})
			return
		}

		if amount < ownership.RentCost*float64(days) {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Amount is lower than rent cost"})
			return
		}

		var rental models.Rental

		rental.UserID = user.(models.User).ID
		rental.OwnerID = ownership.UserID
		rental.TokenID = tokenID
		rental.OwnershipID = ownershipID
		rental.Timestamp = sql.NullTime{Time: time.Now().Add(time.Duration(1e9 * 3600 * 24 * days)), Valid: true}
		rental.Status = "waiting_confirmation"
		rental.TransactionHash = transactionHash
		rental.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
		rental.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

		rentalIDResult, err := ac.rentalRepository.InsertRental(rental)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
			return
		}

		rentalID, err := uuid.Parse(rentalIDResult)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		transaction.RentalID = rentalID
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Transaction type is not valid"})
		return
	}

	token, err := ac.tokenRepository.GetTokenData(tokenID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	transaction.UserFromID = userFromID
	transaction.UserToID = userToID
	transaction.OwnershipID = ownershipID
	transaction.TokenID = tokenID
	transaction.Type = transactionType
	transaction.Quantity = quantity
	transaction.Amount = amount
	transaction.GasFee = gasFee
	transaction.Status = "waiting_confirmation"
	transaction.TransactionHash = transactionHash
	transaction.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	transaction.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	if helpers.IsValidUUID(token.CollectionID) == true {
		transaction.CollectionID = token.CollectionID
	}

	transactionId, err := ac.repository.InsertTransaction(transaction)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	cacheKey := "transaction-list-*"

	var keys []string
	keys, err = ac.redisClient.Keys(cacheKey).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	for _, key := range keys {
		err = ac.redisClient.Del(key).Err()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}
	}

	// Update token's last price
	token.PreviousID = token.ID
	token.NumberOfTransactions = token.NumberOfTransactions + 1
	token.VolumeTransactions = token.VolumeTransactions + amount
	token.LastPrice = amount / float64(quantity)
	token.Status = "waiting_confirmation"
	token.TransactionHash = transactionHash

	_, err = ac.tokenRepository.InsertToken(token, nil)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if helpers.IsValidUUID(token.CollectionID) == true {
		collection, err := ac.collectionRepository.GetCollectionData(token.CollectionID)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
			return
		}

		collection.PreviousID = collection.ID
		collection.NumberOfTransactions = sql.NullInt64{Int64: collection.NumberOfTransactions.Int64 + 1, Valid: true}
		collection.VolumeTransactions = sql.NullFloat64{Float64: collection.VolumeTransactions.Float64 + amount, Valid: true}
		collection.Status = sql.NullString{String: "waiting_confirmation", Valid: true}
		collection.TransactionHash = sql.NullString{String: transactionHash, Valid: true}

		if (amount / float64(quantity)) < collection.Floor.Float64 {
			collection.Floor = sql.NullFloat64{Float64: amount / float64(quantity), Valid: true}
		}

		_, err = ac.collectionRepository.InsertCollection(collection)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
			return
		}
	}

	// Remove token cache
	cacheKey = "token-*"

	keys, err = ac.redisClient.Keys(cacheKey).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
	}

	for _, key := range keys {
		err = ac.redisClient.Del(key).Err()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}
	}

	// Remove rental cache
	cacheKey = "rental-*"

	keys, err = ac.redisClient.Keys(cacheKey).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
	}

	for _, key := range keys {
		err = ac.redisClient.Del(key).Err()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}
	}

	// Remove ownership cache
	cacheKey = "ownership-*"

	keys, err = ac.redisClient.Keys(cacheKey).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
	}

	for _, key := range keys {
		err = ac.redisClient.Del(key).Err()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": gin.H{"transaction_id": transactionId}})
}

func (ac *TransactionController) GetTransactionList(ctx *gin.Context) {
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "25"))

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// Validate user
	userIDParams := ctx.DefaultQuery("user_id", "")
	var userID *uuid.UUID

	if userIDParams != "" {
		userIDConversion, err := uuid.Parse(userIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "User id is not valid"})
			return
		}

		userID = &userIDConversion
	}

	// creatorIDParams := ctx.DefaultQuery("creator_id", "")
	// var creatorID *uuid.UUID

	// if creatorIDParams != "" {
	// 	creatorIDConversion, err := uuid.Parse(creatorIDParams)

	// 	if err != nil {
	// 		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Creator id is not valid"})
	// 		return
	// 	}

	// 	creatorID = &creatorIDConversion
	// }

	// creator := ctx.Query("creator")
	orderBy := ctx.DefaultQuery("order_by", "created_at")
	orderOption := ctx.DefaultQuery("order_option", "ASC")

	// Validate token
	tokenIDParams := ctx.DefaultQuery("token_id", "")
	var tokenID uuid.UUID

	if tokenIDParams != "" {
		tokenID, err = uuid.Parse(tokenIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Token id is not valid"})
			return
		}
	}

	// Validate token
	collectionIDParams := ctx.DefaultQuery("collection_id", "")
	var collectionID uuid.UUID

	if collectionIDParams != "" {
		collectionID, err = uuid.Parse(collectionIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Collection id is not valid"})
			return
		}
	}

	status := "active"

	var transactions []models.Transaction

	cacheKey := fmt.Sprintf("transaction-list-%s-%s-%d-%d-%s-%s-%s", tokenIDParams, collectionIDParams, offset, limit, status, orderBy, orderOption)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &transactions)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"transactions": transactions}})
		return
	}

	if tokenIDParams != "" {
		transactions, err = ac.repository.GetTransactionListByToken(offset, limit, tokenID, &status, orderBy, orderOption)
	} else if collectionIDParams != "" {
		transactions, err = ac.repository.GetTransactionListByCollection(offset, limit, collectionID, &status, orderBy, orderOption)
	} else {
		transactions, err = ac.repository.GetTransactionList(offset, limit, userID, &status, orderBy, orderOption)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheBytes, err := json.Marshal(transactions)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"transactions": transactions}})
}

func (ac *TransactionController) GetTransactionData(ctx *gin.Context) {
	idParam := ctx.Param("id")

	if idParam == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Id parameter is required"})
		return
	}

	id, err := uuid.Parse(idParam)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	var transaction models.Transaction

	cacheKey := fmt.Sprintf("transaction-data-%s", id)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &transaction)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"transaction": transaction}})
		return
	}

	transaction, err = ac.repository.GetTransactionData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if transaction.Amount == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "Transaction not found"})
		return
	}

	cacheBytes, err := json.Marshal(transaction)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"transaction": transaction}})
}

func (ac *TransactionController) UpdateTransaction(ctx *gin.Context) {
	idParam := ctx.Param("id")

	if idParam == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Id parameter is required"})
		return
	}

	id, err := uuid.Parse(idParam)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	transaction, err := ac.repository.GetTransactionData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if transaction.Amount == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "Transaction not found"})
		return
	}

	// Validate quantity
	quantity, err := strconv.Atoi(ctx.DefaultPostForm("quantity", "0"))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if quantity < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Quantity must be greater than 1 or equal"})
		return
	}

	// Validate amount
	amount, err := strconv.ParseFloat(ctx.DefaultPostForm("amount", "0"), 64)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if amount <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Amount must be greater than 0"})
		return
	}

	// Validate gasFee
	gasFee, err := strconv.ParseFloat(ctx.DefaultPostForm("gasFee", "0"), 64)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if gasFee <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Gas fee must be greater than 0"})
		return
	}

	// Validate user from id
	userFromIDParams := ctx.PostForm("user_from_id")
	var userFromID uuid.UUID

	if userFromIDParams != "" {
		userFromID, err = uuid.Parse(userFromIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Category id is not valid"})
			return
		}
	}

	// Validate user to id
	userToIDParams := ctx.PostForm("user_to_id")
	var userToID uuid.UUID

	if userToIDParams != "" {
		userToID, err = uuid.Parse(userToIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Category id is not valid"})
			return
		}
	}

	// Validate token
	tokenIDParams := ctx.PostForm("token_id")
	var tokenID uuid.UUID

	if tokenIDParams != "" {
		tokenID, err = uuid.Parse(tokenIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Token id is not valid"})
			return
		}
	}

	// Validate ownership
	ownershipIDParams := ctx.PostForm("ownership_id")
	var ownershipID uuid.UUID

	if ownershipIDParams != "" {
		ownershipID, err = uuid.Parse(ownershipIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Ownership id is not valid"})
			return
		}
	}

	transaction.UserFromID = userFromID
	transaction.UserToID = userToID
	transaction.OwnershipID = ownershipID
	transaction.TokenID = tokenID
	transaction.Type = ctx.PostForm("type")
	transaction.Quantity = quantity
	transaction.Amount = amount
	transaction.GasFee = gasFee
	transaction.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	err = ac.repository.UpdateTransaction(transaction.ID, transaction)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("transaction-data-%s", id)
	err = ac.redisClient.Del(cacheKey, idParam).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey = "transaction-list-*"

	var keys []string
	keys, err = ac.redisClient.Keys(cacheKey).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	for _, key := range keys {
		err = ac.redisClient.Del(key).Err()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}
	}

	// Remove token cache
	cacheKey = "token-*"

	keys, err = ac.redisClient.Keys(cacheKey).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
	}

	for _, key := range keys {
		err = ac.redisClient.Del(key).Err()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}
	}

	// Remove rental cache
	cacheKey = "rental-*"

	keys, err = ac.redisClient.Keys(cacheKey).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
	}

	for _, key := range keys {
		err = ac.redisClient.Del(key).Err()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}
	}

	// Remove ownership cache
	cacheKey = "ownership-*"

	keys, err = ac.redisClient.Keys(cacheKey).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
	}

	for _, key := range keys {
		err = ac.redisClient.Del(key).Err()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Transaction has been updated"})
}

func (ac *TransactionController) DeleteTransaction(ctx *gin.Context) {
	idParam := ctx.Param("id")

	if idParam == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Id parameter is required"})
		return
	}

	id, err := uuid.Parse(idParam)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.repository.DeleteTransaction(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("transaction-data-%s", id)
	err = ac.redisClient.Del(cacheKey, idParam).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey = "transaction-list-*"

	var keys []string
	keys, err = ac.redisClient.Keys(cacheKey).Result()
	if err != nil {
		panic(err)
	}

	for _, key := range keys {
		err = ac.redisClient.Del(key).Err()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Transaction has been deleted"})
}
