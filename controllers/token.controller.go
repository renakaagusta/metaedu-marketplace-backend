package controllers

import (
	"context"
	"database/sql"
	"encoding/base64"
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

type TokenController struct {
	tokenRepository       *repositories.TokenRepository
	ownershipRepository   *repositories.OwnershipRepository
	collectionRepository  *repositories.CollectionRepository
	transactionRepository *repositories.TransactionRepository
	web3StorageClient     w3s.Client
	redisClient           *redis.Client
}

func NewTokenController(tokenRepository *repositories.TokenRepository, ownershipRepository *repositories.OwnershipRepository, collectionRepository *repositories.CollectionRepository, transactionRepository *repositories.TransactionRepository, web3StorageClient w3s.Client, redisClient *redis.Client) *TokenController {
	return &TokenController{tokenRepository, ownershipRepository, collectionRepository, transactionRepository, web3StorageClient, redisClient}
}

func (ac *TokenController) InsertToken(ctx *gin.Context) {
	// Validate title
	title := ctx.PostForm("title")

	if title == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Title is required"})
		return
	}

	// Validate title
	description := ctx.PostForm("description")

	if description == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Description is required"})
		return
	}

	// Validate title
	var attributes models.Attributes
	var attributesBytes []byte
	attributesParams := ctx.PostForm("attributes")

	if attributesParams != "" {
		err := json.Unmarshal([]byte(attributesParams), &attributes)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err})
			return
		}
	}

	// Validate image
	var imageFileName string
	uploadedImage, err := ctx.FormFile("image")

	if err != nil {
		uploadedImageBase64 := ctx.PostForm("image")
		uploadedImageName := ctx.PostForm("image_name")

		if uploadedImageBase64 == "" || uploadedImageName == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Image is required"})
			return
		}

		uploadedImageBase64Decode, err := base64.StdEncoding.DecodeString(uploadedImageBase64)
		if err != nil {
			panic(err)
		}

		file, err := os.Create(uploadedImageName)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		defer file.Close()

		if _, err := file.Write(uploadedImageBase64Decode); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		if err := file.Sync(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		imageFileName = file.Name()
	} else {
		if err := ctx.SaveUploadedFile(uploadedImage, uploadedImage.Filename); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		imageFileName = uploadedImage.Filename
	}

	// Validate supply
	supply, err := strconv.Atoi(ctx.DefaultPostForm("supply", "0"))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if supply < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Supply must be greater than 1 or equal"})
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

	// Check if user is exist in request
	user, isExist := ctx.Get("user")

	if !isExist {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": "User data is not valid"})
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

	// Validate collection
	collectionIDParams := ctx.PostForm("collection_id")
	var collectionID uuid.UUID

	if collectionIDParams != "" {
		collectionID, err = uuid.Parse(collectionIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Collection id is not valid"})
			return
		}

		collection, err := ac.collectionRepository.GetCollectionData(collectionID)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err})
			return
		}

		if collection.CreatorID != user.(models.User).ID {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "User has no access to add token to this collection"})
			return
		}

		collection.Status = sql.NullString{String: "waiting_confirmation", Valid: true}
		collection.NumberOfItems = sql.NullInt64{Int64: collection.NumberOfItems.Int64 + 1, Valid: true}
		collection.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

		err = ac.collectionRepository.UpdateCollection(collectionID, collection)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err})
			return
		}
	}

	// Validate fraction
	fractionIDParams := ctx.PostForm("fraction_id")
	var fractionID uuid.UUID

	if fractionIDParams != "" {
		fractionID, err = uuid.Parse(fractionIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Fraction id is not valid"})
			return
		}
	}

	imageFile, err := os.Open(imageFileName)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	imageCid, err := ac.web3StorageClient.Put(context.Background(), imageFile)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	imageUrl := fmt.Sprintf("https://%s.ipfs.dweb.link/%s", imageCid.String(), imageFileName)

	err = os.Remove(imageFileName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	tokenIndex, err := ac.tokenRepository.GetLastTokenIndex()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	var token models.Token

	token.TokenIndex = tokenIndex + 1
	token.Title = title
	token.Description = description
	token.Supply = supply
	token.LastPrice = price
	token.InitialPrice = price
	token.Image = imageUrl
	token.Views = 0
	token.NumberOfTransactions = 0
	token.VolumeTransactions = 0
	token.CreatorID = user.(models.User).ID
	token.Attributes = attributes
	token.Status = "waiting_confirmation"
	token.TransactionHash = ""
	token.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	token.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	if categoryIDParams != "" {
		token.CategoryID = categoryID
	}

	if collectionIDParams != "" {
		token.CollectionID = collectionID
	}

	if fractionIDParams != "" {
		token.FractionID = fractionID
	}

	tokenJson, err := json.Marshal(&token)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	tmpUriFileName := fmt.Sprint(time.Now().UnixNano()/int64(time.Millisecond)) + ".json"
	tmpUriFile, err := os.Create(tmpUriFileName)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	defer tmpUriFile.Close()

	_, err = tmpUriFile.WriteString(string(tokenJson))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	uriFile, err := os.Open(tmpUriFileName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	uriFileName := path.Base(tmpUriFileName)

	uriCid, err := ac.web3StorageClient.Put(context.Background(), uriFile)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	uriUrl := fmt.Sprintf("https://%s.ipfs.dweb.link/%s", uriCid.String(), uriFileName)

	err = os.Remove(uriFileName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	token.Uri = uriUrl

	token, err = ac.tokenRepository.InsertToken(token, attributesBytes)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	// Remove token cache
	cacheKey := "token-list-*"

	var keys []string
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

	var ownership models.Ownership

	ownership.TokenID = token.ID
	ownership.UserID = user.(models.User).ID
	ownership.Quantity = supply
	ownership.SalePrice = price
	ownership.RentCost = price
	ownership.AvailableForRent = false
	ownership.AvailableForSale = false
	ownership.Status = "waiting_confirmation"
	ownership.TransactionHash = ""
	ownership.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	ownership.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	ownershipIDResult, err := ac.ownershipRepository.InsertOwnership(ownership)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
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

	// Remove collection cache
	cacheKey = "collection-*"

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

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": gin.H{"token": token, "ownership_id": ownershipIDResult}})
}

func (ac *TokenController) GetTokenList(ctx *gin.Context) {
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

	// Validate min price
	minPrice, err := strconv.Atoi(ctx.DefaultQuery("min_price", "0"))

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// Validate max price
	maxPrice, err := strconv.Atoi(ctx.DefaultQuery("max_price", "0"))

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// Validate category id
	categoryIDParams := ctx.DefaultQuery("category_id", "")
	var categoryID *uuid.UUID

	if categoryIDParams != "" {
		categoryIDConversion, err := uuid.Parse(categoryIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Category id is not valid"})
			return
		}

		categoryID = &categoryIDConversion
	}

	// Validate collection id
	collectionIDParams := ctx.DefaultQuery("collection_id", "")
	var collectionID *uuid.UUID

	if collectionIDParams != "" {
		collectionIDConversion, err := uuid.Parse(collectionIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Category id is not valid"})
			return
		}

		collectionID = &collectionIDConversion
	}

	creatorIDParams := ctx.DefaultQuery("creator_id", "")
	var creatorID *uuid.UUID

	if creatorIDParams != "" {
		creatorIDConversion, err := uuid.Parse(creatorIDParams)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Category id is not valid"})
			return
		}

		creatorID = &creatorIDConversion
	}

	keyword := ctx.DefaultQuery("keyword", "")
	orderBy := ctx.DefaultQuery("order_by", "created_at")
	orderOption := ctx.DefaultQuery("order_option", "ASC")

	var tokens []models.Token

	cacheKey := fmt.Sprintf("token-list-%d-%d-%s-%s-%s-%d-%d-%s-%s-%s", offset, limit, keyword, orderBy, orderOption, minPrice, maxPrice, categoryIDParams, collectionIDParams, creatorIDParams)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &tokens)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"tokens": tokens}})
		return
	}

	status := "active"

	tokens, err = ac.tokenRepository.GetTokenList(offset, limit, keyword, categoryID, collectionID, creatorID, minPrice, maxPrice, &status, orderBy, orderOption)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheBytes, err := json.Marshal(tokens)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"tokens": tokens}})
}

func (ac *TokenController) GetTokenData(ctx *gin.Context) {
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

	var token models.Token

	cacheKey := fmt.Sprintf("token-data-%s", id)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &token)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		token.Views = token.Views + 1

		err = ac.tokenRepository.UpdateToken(token.ID, token)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		cacheBytes, err := json.Marshal(token)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"token": token}})
		return
	}

	token, err = ac.tokenRepository.GetTokenData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if token.Uri == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "Token not found"})
		return
	}

	cacheBytes, err := json.Marshal(token)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	token.Views = token.Views + 1

	err = ac.tokenRepository.UpdateToken(token.ID, token)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"token": token}})
}

func (ac *TokenController) UpdateToken(ctx *gin.Context) {
	// Validate transaction hash
	transactionHash := ctx.DefaultPostForm("transaction_hash", "")

	if transactionHash == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Transaction hash is required"})
		return
	}

	// Validate id
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

	// Validate user
	user, isExist := ctx.Get("user")

	if !isExist {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": "User data is not valid"})
		return
	}

	token, err := ac.tokenRepository.GetTokenData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	// Check if token is found
	if token.Uri == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "Token not found"})
		return
	}

	// Check if token has been updated
	if token.TransactionHash != "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Token has been updated"})
		return
	}

	// Check if user is creator
	if token.CreatorID != user.(models.User).ID {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Only creator can update token`s transaction hash"})
		return
	}

	// Update user transaction hash
	token.TransactionHash = transactionHash
	token.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	err = ac.tokenRepository.UpdateToken(token.ID, token)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	// Get ownership data
	ownership, err := ac.ownershipRepository.GetOwnershipByTokenAndUser(token.ID, user.(models.User).ID)

	// Update ownership transaction hash
	ownership.TransactionHash = transactionHash

	err = ac.ownershipRepository.UpdateOwnership(ownership.ID, ownership)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	// Remove token cache
	cacheKey := "token-list-*"

	var keys []string
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

	// Remove collection cache
	cacheKey = "collection-*"

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

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": gin.H{"token": token}})
}

func (ac *TokenController) DeleteToken(ctx *gin.Context) {
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

	err = ac.tokenRepository.DeleteToken(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("token-data-%s", id)
	err = ac.redisClient.Del(cacheKey, idParam).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey = "token-list-*"

	var keys []string
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

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Token has been deleted"})
}

func (ac *TokenController) GetTokenTransactionList(ctx *gin.Context) {
	// Validate token id
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
	transactions, err = ac.transactionRepository.GetTransactionListByToken(limit, offset, id, &status, orderBy, orderOption)

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
