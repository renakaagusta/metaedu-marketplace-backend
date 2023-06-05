package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	models "metaedu-marketplace/models"
	"metaedu-marketplace/repositories"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/web3-storage/go-w3s-client"
)

type CollectionController struct {
	repository            *repositories.CollectionRepository
	transactionRepository *repositories.TransactionRepository
	web3StorageClient     w3s.Client
	redisClient           *redis.Client
}

func NewCollectionController(repository *repositories.CollectionRepository, transactionRepository *repositories.TransactionRepository, web3StorageClient w3s.Client, redisClient *redis.Client) *CollectionController {
	return &CollectionController{repository, transactionRepository, web3StorageClient, redisClient}
}

func (ac *CollectionController) InsertCollection(ctx *gin.Context) {
	uploadedThumbnail, err := ctx.FormFile("thumbnail")

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Thumbnail is required"})
		return
	}

	if err := ctx.SaveUploadedFile(uploadedThumbnail, uploadedThumbnail.Filename); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	uploadedCover, err := ctx.FormFile("cover")

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Cover is required"})
		return
	}

	if err := ctx.SaveUploadedFile(uploadedCover, uploadedCover.Filename); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// Validate creator
	user, isExist := ctx.Get("user")

	if !isExist {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": "User data is not valid"})
		return
	}

	thumbnailFile, err := os.Open(uploadedThumbnail.Filename)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	thumbnailFileName := path.Base(uploadedThumbnail.Filename)

	thumbnailCid, err := ac.web3StorageClient.Put(context.Background(), thumbnailFile)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	thumbnailUrl := fmt.Sprintf("https://%s.ipfs.dweb.link/%s\n", thumbnailCid.String(), thumbnailFileName)

	err = os.Remove(thumbnailFileName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	coverFile, err := os.Open(uploadedCover.Filename)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	coverFileName := path.Base(uploadedCover.Filename)

	coverCid, err := ac.web3StorageClient.Put(context.Background(), coverFile)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	coverUrl := fmt.Sprintf("https://%s.ipfs.dweb.link/%s\n", coverCid.String(), coverFileName)

	err = os.Remove(coverFileName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// Validate category
	categoryIDParams := ctx.PostForm("category_id")
	var categoryID uuid.UUID

	if categoryIDParams != "" {
		categoryID, err = uuid.Parse(categoryIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Category id is not valid"})
			return
		}
	}

	var collection models.Collection

	collection.Title = sql.NullString{String: ctx.PostForm("title"), Valid: true}
	collection.Views = sql.NullInt64{Int64: 0, Valid: true}
	collection.NumberOfItems = sql.NullInt64{Int64: 0, Valid: true}
	collection.NumberOfTransactions = sql.NullInt64{Int64: 0, Valid: true}
	collection.VolumeTransactions = sql.NullFloat64{Float64: 0, Valid: true}
	collection.Description = sql.NullString{String: ctx.PostForm("default"), Valid: true}
	collection.Thumbnail = sql.NullString{String: thumbnailUrl, Valid: true}
	collection.Cover = sql.NullString{String: coverUrl, Valid: true}
	collection.CreatorID = user.(models.User).ID
	collection.CategoryID = categoryID
	collection.Status = sql.NullString{String: "active", Valid: true}
	collection.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	collection.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	collectionId, err := ac.repository.InsertCollection(collection)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	cacheKey := "collection-list-*"

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

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": gin.H{"collectionId": collectionId}})
}

func (ac *CollectionController) GetCollectionList(ctx *gin.Context) {
	keyword := ctx.DefaultQuery("keyword", "")
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

	orderBy := ctx.DefaultQuery("order_by", "created_at")
	orderOption := ctx.DefaultQuery("order_option", "ASC")

	creatorIDParams := ctx.DefaultQuery("creator_id", "")
	var creatorID *uuid.UUID

	if creatorIDParams != "" {
		creatorIDConversion, err := uuid.Parse(creatorIDParams)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		creatorID = &creatorIDConversion
	}

	status := "active"

	var collections []models.Collection

	cacheKey := fmt.Sprintf("collection-list-%s-%d-%d-%s-%s-%s", creatorIDParams, offset, limit, status, orderBy, orderOption)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &collections)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"collections": collections}})
		return
	}

	collections, err = ac.repository.GetCollectionList(offset, limit, keyword, creatorID, &status, orderBy, orderOption)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheBytes, err := json.Marshal(collections)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"collections": collections}})
}

func (ac *CollectionController) GetCollectionData(ctx *gin.Context) {
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

	var collection models.Collection

	cacheKey := fmt.Sprintf("collection-data-%s", id)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &collection)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		collection.Views = sql.NullInt64{Int64: collection.Views.Int64 + 1, Valid: true}

		err = ac.repository.UpdateCollection(collection.ID, collection)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"collection": collection}})
		return
	}

	collection, err = ac.repository.GetCollectionData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if collection.Title.String == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "Collection not found"})
		return
	}

	cacheBytes, err := json.Marshal(collection)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	collection.Views = sql.NullInt64{Int64: collection.Views.Int64 + 1, Valid: true}

	err = ac.repository.UpdateCollection(collection.ID, collection)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"collection": collection}})
}

func (ac *CollectionController) UpdateCollection(ctx *gin.Context) {
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

	collection, err := ac.repository.GetCollectionData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if collection.Title.String == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "Collection not found"})
		return
	}

	// Validate creator
	creatorIDParams := ctx.PostForm("creator_id")
	var creatorID uuid.UUID

	if creatorIDParams != "" {
		creatorID, err = uuid.Parse(creatorIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Owner id is not valid"})
			return
		}
	}

	// Validate category
	categoryIDParams := ctx.PostForm("category_id")
	var categoryID uuid.UUID

	if categoryIDParams != "" {
		categoryID, err = uuid.Parse(categoryIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Category id is not valid"})
			return
		}
	}

	collection.Title = sql.NullString{String: ctx.DefaultPostForm("title", collection.Title.String), Valid: true}
	collection.Description = sql.NullString{String: ctx.DefaultPostForm("description", collection.Description.String), Valid: true}
	collection.CategoryID = categoryID
	collection.CreatorID = creatorID
	collection.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	err = ac.repository.UpdateCollection(collection.ID, collection)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("collection-data-%s", id)
	err = ac.redisClient.Del(cacheKey, idParam).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey = "collection-list-*"

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

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Collection has been updated"})
}

func (ac *CollectionController) DeleteCollection(ctx *gin.Context) {
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

	err = ac.repository.DeleteCollection(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("collection-data-%s", id)
	err = ac.redisClient.Del(cacheKey, idParam).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey = "collection-list-*"

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

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Collection has been deleted"})
}

func (ac *CollectionController) GetCollectionTransactionList(ctx *gin.Context) {
	// Validate collection id
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

	// Validate limit
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "25"))

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// Validate offset
	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	keyword := ctx.DefaultQuery("keyword", "")
	orderBy := ctx.DefaultQuery("order_by", "created_at")
	orderOption := ctx.DefaultQuery("order_option", "ASC")

	status := "active"

	var transactions []models.Transaction

	// Get cache data
	cacheKey := fmt.Sprintf("transaction-list-%s-%d-%d-%s-%s-%s", id.String(), offset, limit, keyword, orderBy, orderOption)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// If cache is exist return data from cache
	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &transactions)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"transactions": transactions}})
		return
	}

	// Get transaction list data from repository
	transactions, err = ac.transactionRepository.GetTransactionListByCollection(offset, limit, id, &status, orderBy, orderOption)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// Set transaction list data from repository to redis
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
