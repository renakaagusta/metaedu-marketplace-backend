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

type RentalController struct {
	repository        *repositories.RentalRepository
	web3StorageClient w3s.Client
	redisClient       *redis.Client
}

func NewRentalController(repository *repositories.RentalRepository, web3StorageClient w3s.Client, redisClient *redis.Client) *RentalController {
	return &RentalController{repository, web3StorageClient, redisClient}
}

func (ac *RentalController) InsertRental(ctx *gin.Context) {
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

	// Validate owner
	ownerIDParams := ctx.PostForm("owner_id")
	var ownerID uuid.UUID

	if ownerIDParams != "" {
		ownerID, err = uuid.Parse(ownerIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "User id is not valid"})
			return
		}
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "User id is required"})
		return
	}

	// Validate ownership
	ownershipIDParams := ctx.PostForm("ownership_id")
	var ownershipID uuid.UUID

	if ownershipIDParams != "" {
		ownershipID, err = uuid.Parse(ownershipIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "User id is not valid"})
			return
		}
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "User id is required"})
		return
	}

	// Validate days
	daysParams := ctx.PostForm("days")

	if daysParams == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Timestamp is required"})
		return
	}

	days, err := strconv.Atoi(daysParams)

	var rental models.Rental

	rental.UserID = userID
	rental.OwnerID = ownerID
	rental.TokenID = tokenID
	rental.OwnershipID = ownershipID
	rental.Timestamp = sql.NullTime{Time: time.Now().Add(time.Duration(1e9 * 3600 * 24 * days)), Valid: true}
	rental.Status = "waiting_confirmation"
	rental.TransactionHash = transactionHash
	rental.UpdatedAt = sql.NullTime{}
	rental.CreatedAt = sql.NullTime{}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	rentalId, err := ac.repository.InsertRental(rental)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	cacheKey := "rental-list-*"

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

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": gin.H{"rental_id": rentalId}})
}

func (ac *RentalController) GetRentalList(ctx *gin.Context) {
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

	userIDParams := ctx.DefaultQuery("user_id", "")
	userID, err := uuid.Parse(userIDParams)

	keyword := ctx.DefaultQuery("keyword", "")
	orderBy := ctx.DefaultQuery("order_by", "created_at")
	orderOption := ctx.DefaultQuery("order_option", "ASC")

	var rentals []models.Rental

	cacheKey := fmt.Sprintf("rental-list-%d-%d-%s-%s-%s-%s", offset, limit, keyword, orderBy, orderOption, userID)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &rentals)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"rentals": rentals}})
		return
	}

	status := "active"

	if userIDParams != "" {
		rentals, err = ac.repository.GetRentalListByUserID(offset, limit, userID, orderBy, orderOption, &status)
	} else {
		rentals, err = ac.repository.GetRentalList(offset, limit, &status, orderBy, orderOption)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheBytes, err := json.Marshal(rentals)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"rentals": rentals}})
}

func (ac *RentalController) GetRentalData(ctx *gin.Context) {
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

	var rental models.Rental

	cacheKey := fmt.Sprintf("rental-data-%s", id)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &rental)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"rental": rental}})
		return
	}

	rental, err = ac.repository.GetRentalData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// if rental.Quantity == 0 {
	// 	ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "Rental not found"})
	// 	return
	// }

	cacheBytes, err := json.Marshal(rental)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"rental": rental}})
}

func (ac *RentalController) UpdateRental(ctx *gin.Context) {
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

	rental, err := ac.repository.GetRentalData(id)

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

	// Validate ownership
	ownershipIDParams := ctx.PostForm("ownership_id")
	var ownershipID uuid.UUID

	if ownershipIDParams != "" {
		ownershipID, err = uuid.Parse(ownershipIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "User id is not valid"})
			return
		}
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "User id is required"})
		return
	}

	// // Validate days
	// daysParams := ctx.PostForm("days")

	// if daysParams == "" {
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Timestamp is required"})
	// 	return
	// }

	// days, err := strconv.Atoi(daysParams)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	rental.UserID = userID
	rental.TokenID = tokenID
	rental.OwnershipID = ownershipID
	// rental.Timestamp = sql.NullTime{}.Time .Now().Add(time.Duration(1e9 * 3600 * 24 * days))
	rental.UpdatedAt = sql.NullTime{}

	err = ac.repository.UpdateRental(rental.ID, rental)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("rental-data-%s", id)
	err = ac.redisClient.Del(cacheKey, idParam).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey = "rental-list-*"

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

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Rental has been updated"})
}

func (ac *RentalController) DeleteRental(ctx *gin.Context) {
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

	err = ac.repository.DeleteRental(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("rental-data-%s", id)
	err = ac.redisClient.Del(cacheKey, idParam).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey = "rental-list-*"

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

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Rental has been deleted"})
}
