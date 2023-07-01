package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	models "metaedu-marketplace/models"
	"metaedu-marketplace/repositories"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/web3-storage/go-w3s-client"
)

type FractionController struct {
	repository          *repositories.FractionRepository
	tokenRepository     *repositories.TokenRepository
	ownershipRepository *repositories.OwnershipRepository
	rentalRepository    *repositories.RentalRepository
	userRepository      *repositories.UserRepository
	web3StorageClient   w3s.Client
	redisClient         *redis.Client
}

func NewFractionController(repository *repositories.FractionRepository, tokenRepository *repositories.TokenRepository, ownershipRepository *repositories.OwnershipRepository, rentalRepository *repositories.RentalRepository, userRepository *repositories.UserRepository, web3StorageClient w3s.Client, redisClient *redis.Client) *FractionController {
	return &FractionController{repository, tokenRepository, ownershipRepository, rentalRepository, userRepository, web3StorageClient, redisClient}
}

func (ac *FractionController) InsertFraction(ctx *gin.Context) {
	var err error

	// Validate transaction hash
	transactionHash := ctx.DefaultPostForm("transaction_hash", "")

	if transactionHash == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Transaction hash is required"})
		return
	}

	// Validate token parent id
	tokenSourceIDParams := ctx.PostForm("token_source_id")
	var tokenSourceID uuid.UUID

	tokenSourceID, err = uuid.Parse(tokenSourceIDParams)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Token source id is not valid"})
		return
	}

	// Validate ownership id
	ownershipIDParams := ctx.PostForm("ownership_id")
	var ownershipID uuid.UUID

	ownershipID, err = uuid.Parse(ownershipIDParams)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Ownership id is not valid"})
		return
	}

	// Validate supply
	supply, err := strconv.Atoi(ctx.DefaultPostForm("supply", "0"))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if supply < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Quantity must be greater than 1 or equal"})
		return
	}

	// Validate price
	price, err := strconv.ParseFloat(ctx.DefaultPostForm("price", "0"), 64)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if price <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Price must be greater than 0"})
		return
	}

	// Default status for query data
	status := "active"

	// Check if token is still in rental period
	rentals, err := ac.rentalRepository.GetRentalList(100, 0, "", nil, nil, nil, nil, nil, nil, &tokenSourceID, &status, "created_at", "DESC")

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	for _, rental := range rentals {
		if rental.Timestamp.Time.After(time.Now()) {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Item is still in rental period"})
			return
		}
	}

	// Get last token index
	tokenIndex, err := ac.tokenRepository.GetLastTokenIndex()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	tokenSource, err := ac.tokenRepository.GetTokenData(tokenSourceID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	var tokenFraction models.Token

	tokenFraction.SourceID = tokenSourceID
	tokenFraction.TokenIndex = tokenIndex + 1
	tokenFraction.Title = tokenSource.Title
	tokenFraction.Description = tokenSource.Description
	tokenFraction.CategoryID = tokenSource.CategoryID
	tokenFraction.CollectionID = tokenSource.CollectionID
	tokenFraction.Supply = supply
	tokenFraction.LastPrice = price
	tokenFraction.InitialPrice = price
	tokenFraction.Image = tokenSource.Image
	tokenFraction.Uri = tokenSource.Uri
	tokenFraction.Views = 0
	tokenFraction.NumberOfTransactions = 0
	tokenFraction.VolumeTransactions = 0
	tokenFraction.Status = "waiting_confirmation"
	tokenFraction.TransactionHash = transactionHash
	tokenFraction.CreatorID = tokenSource.CreatorID
	tokenFraction.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	tokenFraction.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	tokenFractionUpdated, err := ac.tokenRepository.InsertToken(tokenFraction, nil)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	cacheKey := "token-list-*"

	var keys []string
	keys, err = ac.redisClient.Keys(cacheKey).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	tokenSource.PreviousID = tokenSource.ID
	tokenSource.FractionID = tokenFractionUpdated.ID
	tokenSource.Status = "waiting_confirmation"
	tokenSource.TransactionHash = transactionHash

	_, err = ac.tokenRepository.InsertToken(tokenSource, nil)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	cacheKey = fmt.Sprintf("token-data-%s", tokenSourceID)
	err = ac.redisClient.Del(cacheKey).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// Get system user
	systemUser, err := ac.userRepository.GetSystemUser()

	// Get ownership data of parent token
	oldOwnership, err := ac.ownershipRepository.GetOwnershipData(ownershipID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	oldOwnership.UserID = systemUser.ID
	oldOwnership.PreviousID = oldOwnership.ID
	oldOwnership.Status = "waiting_confirmation"
	oldOwnership.TransactionHash = transactionHash

	_, err = ac.ownershipRepository.InsertOwnership(oldOwnership)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err})
		return
	}

	cacheKey = fmt.Sprintf("ownership-data-%s", ownershipID)
	err = ac.redisClient.Del(cacheKey).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// Create ownership data for fraction token
	var newOwnership models.Ownership

	newOwnership.TokenID = tokenFractionUpdated.ID
	newOwnership.UserID = tokenFraction.CreatorID
	newOwnership.Quantity = supply
	newOwnership.SalePrice = price
	newOwnership.RentCost = oldOwnership.RentCost
	newOwnership.AvailableForRent = false
	newOwnership.AvailableForSale = false
	newOwnership.Status = "waiting_confirmation"
	newOwnership.TransactionHash = transactionHash
	newOwnership.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	newOwnership.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	ownershipIDResult, err := ac.ownershipRepository.InsertOwnership(newOwnership)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey = "ownership-list-*"

	keys, err = ac.redisClient.Keys(cacheKey).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// Create fraction data
	var fraction models.Fraction

	fraction.TokenParentID = tokenSourceID
	fraction.TokenFractionID = tokenFraction.ID
	fraction.Status = "waiting_confirmation"
	fraction.TransactionHash = transactionHash
	fraction.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	fraction.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	fractionId, err := ac.repository.InsertFraction(fraction)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	// Remove fraction cache
	cacheKey = "fraction-list-*"
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

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": gin.H{"fraction_id": fractionId, "ownership_id": ownershipIDResult, "token_id": tokenFractionUpdated.ID}})
}

func (ac *FractionController) GetFractionList(ctx *gin.Context) {
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

	keyword := ctx.DefaultQuery("keywprd", "")
	creator := ctx.DefaultQuery("creator", "")
	creatorIDParams := ctx.DefaultQuery("creator_id", "")

	var creatorID *uuid.UUID

	if creatorIDParams != "" {
		creatorIDConversion, err := uuid.Parse(creatorIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Creator id is not valid"})
			return
		}

		creatorID = &creatorIDConversion
	}

	orderBy := ctx.DefaultQuery("order_by", "created_at")
	orderOption := ctx.DefaultQuery("order_option", "ASC")

	status := "active"

	var fractions []models.Fraction

	cacheKey := fmt.Sprintf("fraction-list-%d-%d-%s-%s-%s-%s-%s", offset, limit, creatorID, creator, status, orderBy, orderOption)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &fractions)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"fractions": fractions}})
		return
	}

	fractions, err = ac.repository.GetFractionList(offset, limit, keyword, creatorID, &creator, &status, orderBy, orderOption)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheBytes, err := json.Marshal(fractions)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"fractions": fractions}})
}

func (ac *FractionController) GetFractionData(ctx *gin.Context) {
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

	var fraction models.Fraction

	cacheKey := fmt.Sprintf("fraction-data-%s", id)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &fraction)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"fraction": fraction}})
		return
	}

	fraction, err = ac.repository.GetFractionData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// if fraction.Amount == 0 {
	// 	ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "Fraction not found"})
	// 	return
	// }

	cacheBytes, err := json.Marshal(fraction)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"fraction": fraction}})
}

func (ac *FractionController) UpdateFraction(ctx *gin.Context) {
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

	fraction, err := ac.repository.GetFractionData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// if fraction.Amount == 0 {
	// 	ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "Fraction not found"})
	// 	return
	// }

	// Validate token parent id
	tokenSourceIDParams := ctx.PostForm("token_source_id")
	var tokenSourceID uuid.UUID

	tokenSourceID, err = uuid.Parse(tokenSourceIDParams)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Category id is not valid"})
		return
	}

	// Validate token fraction id
	tokenFractionIdParams := ctx.PostForm("user_to_id")
	var tokenFractionId uuid.UUID

	if tokenFractionIdParams != "" {
		tokenFractionId, err = uuid.Parse(tokenFractionIdParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Category id is not valid"})
			return
		}
	}

	fraction.TokenParentID = tokenSourceID
	fraction.TokenFractionID = tokenFractionId
	fraction.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	err = ac.repository.UpdateFraction(fraction.ID, fraction)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("fraction-data-%s", id)
	err = ac.redisClient.Del(cacheKey, idParam).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey = "fraction-list-*"

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

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Fraction has been updated"})
}

func (ac *FractionController) DeleteFraction(ctx *gin.Context) {
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

	err = ac.repository.DeleteFraction(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("fraction-data-%s", id)
	err = ac.redisClient.Del(cacheKey, idParam).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey = "fraction-list-*"

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

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Fraction has been deleted"})
}
