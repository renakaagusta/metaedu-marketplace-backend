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

type OwnershipController struct {
	repository        *repositories.OwnershipRepository
	tokenRepository   *repositories.TokenRepository
	rentalRepository  *repositories.RentalRepository
	web3StorageClient w3s.Client
	redisClient       *redis.Client
}

func NewOwnershipController(repository *repositories.OwnershipRepository, tokenRepository *repositories.TokenRepository, rentalRepository *repositories.RentalRepository, web3StorageClient w3s.Client, redisClient *redis.Client) *OwnershipController {
	return &OwnershipController{repository, tokenRepository, rentalRepository, web3StorageClient, redisClient}
}

func (ac *OwnershipController) InsertOwnership(ctx *gin.Context) {
	// Validate transaction hash
	transactionHash := ctx.DefaultPostForm("transaction_hash", "")

	if transactionHash == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Transaction hash is required"})
		return
	}

	// Validate token
	tokenIDParams := ctx.PostForm("token_id")
	var tokenID uuid.UUID
	var err error

	if tokenIDParams != "" {
		tokenID, err = uuid.Parse(tokenIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Token id is not valid"})
			return
		}
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Token id is required"})
		return
	}

	// Validate user
	userIDParams := ctx.PostForm("user_id")
	var userID uuid.UUID

	if userIDParams != "" {
		userID, err = uuid.Parse(userIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "User id is not valid"})
			return
		}
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "User id is required"})
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

	// Validate sale price
	salePrice, err := strconv.ParseFloat(ctx.DefaultPostForm("sale_price", "0"), 64)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if salePrice <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Sale price must be greater than 0"})
		return
	}

	// Validate rent cost
	rentCost, err := strconv.ParseFloat(ctx.DefaultPostForm("rent_cost", "0"), 64)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if rentCost <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Sale price must be greater than 0"})
		return
	}

	availableForSale := ctx.DefaultPostForm("available_for_sale", "false")
	availableForRent := ctx.DefaultPostForm("available_for_rent", "false")

	var ownership models.Ownership

	ownership.UserID = userID
	ownership.TokenID = tokenID
	ownership.Quantity = quantity
	ownership.SalePrice = salePrice
	ownership.RentCost = rentCost
	ownership.AvailableForSale = availableForSale == "true"
	ownership.AvailableForRent = availableForRent == "true"
	ownership.Status = "waiting_confirmation"
	ownership.TransactionHash = transactionHash
	ownership.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	ownership.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	ownershipId, err := ac.repository.InsertOwnership(ownership)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	cacheKey := "ownership-list-*"

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

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": gin.H{"ownership_id": ownershipId}})
}

func (ac *OwnershipController) GetOwnershipList(ctx *gin.Context) {
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

	keyword := ctx.DefaultQuery("keyword", "")
	user := ctx.DefaultQuery("user", "")
	userIDParams := ctx.DefaultQuery("user_id", "")
	creator := ctx.DefaultQuery("creator", "")
	creatorIDParams := ctx.DefaultQuery("creator_id", "")
	tokenIDParams := ctx.DefaultQuery("token_id", "")
	orderBy := ctx.DefaultQuery("order_by", "created_at")
	orderOption := ctx.DefaultQuery("order_option", "ASC")

	if (userIDParams == "" && tokenIDParams == "") && (user == "" && creator == "") {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "User id/user/token id/creator are required"})
		return
	}

	var userID *uuid.UUID

	if userIDParams != "" {
		userIDConversion, err := uuid.Parse(userIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "User id is not valid"})
			return
		}

		userID = &userIDConversion
	}

	var creatorID *uuid.UUID

	if creatorIDParams != "" {
		creatorIDConversion, err := uuid.Parse(creatorIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Creator id is not valid"})
			return
		}

		creatorID = &creatorIDConversion
	}

	var tokenID *uuid.UUID

	if tokenIDParams != "" {
		tokenIDConversion, err := uuid.Parse(tokenIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Token id is not valid"})
			return
		}

		tokenID = &tokenIDConversion
	}

	var ownerships []models.Ownership

	cacheKey := fmt.Sprintf("ownership-list-%d-%d-%s-%s-%s-%s-%s-%s-%s", offset, limit, keyword, userIDParams, user, creatorIDParams, creator, tokenIDParams, orderBy, orderOption)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &ownerships)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"ownerships": ownerships}})
		return
	}

	ownerships, err = ac.repository.GetOwnershipList(offset, limit, keyword, userID, &user, creatorID, &creator, tokenID, "active", orderBy, orderOption)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheBytes, err := json.Marshal(ownerships)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"ownerships": ownerships}})
}

func (ac *OwnershipController) GetOwnershipData(ctx *gin.Context) {
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

	var ownership models.Ownership

	cacheKey := fmt.Sprintf("ownership-data-%s", id)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &ownership)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"ownership": ownership}})
		return
	}

	ownership, err = ac.repository.GetOwnershipData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if ownership.Quantity == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "Ownership not found"})
		return
	}

	cacheBytes, err := json.Marshal(ownership)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"ownership": ownership}})
}

func (ac *OwnershipController) UpdateOwnership(ctx *gin.Context) {
	// Validate transaction hash
	transactionHash := ctx.DefaultPostForm("transaction_hash", "")

	if transactionHash == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Transaction hash is required"})
		return
	}

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

	ownership, err := ac.repository.GetOwnershipData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	token, err := ac.tokenRepository.GetTokenData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
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
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Token id is required"})
		return
	}

	// Validate user
	userIDParams := ctx.PostForm("user_id")
	var userID uuid.UUID

	if userIDParams != "" {
		userID, err = uuid.Parse(userIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "User id is not valid"})
			return
		}
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "User id is required"})
		return
	}

	// Validate sale price
	salePrice, err := strconv.ParseFloat(ctx.DefaultPostForm("sale_price", "0"), 64)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if salePrice <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Sale price must be greater than 0"})
		return
	}

	// Validate rent cost
	rentCost, err := strconv.ParseFloat(ctx.DefaultPostForm("rent_cost", "0"), 64)

	if token.Supply == 1 {
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
			return
		}

		if rentCost <= 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Rent cost must be greater than 0"})
			return
		}
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
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Item is still in rental period"})
			return
		}
	}

	availableForSale := ctx.DefaultPostForm("available_for_sale", "false")
	availableForRent := ctx.DefaultPostForm("available_for_rent", "false")

	ownership.PreviousID = ownership.ID
	ownership.UserID = userID
	ownership.TokenID = tokenID
	ownership.SalePrice = salePrice
	ownership.RentCost = rentCost
	ownership.AvailableForSale = availableForSale == "true"
	ownership.AvailableForRent = availableForRent == "true"
	ownership.Status = "waiting_confirmation"
	ownership.TransactionHash = transactionHash
	ownership.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	_, err = ac.repository.InsertOwnership(ownership)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("ownership-data-%s", id)
	err = ac.redisClient.Del(cacheKey, idParam).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey = "ownership-list-*"

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

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Ownership has been updated"})
}

func (ac *OwnershipController) DeleteOwnership(ctx *gin.Context) {
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

	err = ac.repository.DeleteOwnership(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("ownership-data-%s", id)
	err = ac.redisClient.Del(cacheKey, idParam).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey = "ownership-list-*"

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

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Ownership has been deleted"})
}
