package controllers

import (
	"context"
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
	repository        *repositories.TokenRepository
	web3StorageClient w3s.Client
	redisClient       *redis.Client
}

func NewTokenController(repository *repositories.TokenRepository, web3StorageClient w3s.Client, redisClient *redis.Client) *TokenController {
	return &TokenController{repository, web3StorageClient, redisClient}
}

func (ac *TokenController) InsertToken(ctx *gin.Context) {
	uploadedFile, err := ctx.FormFile("image")

	if err := ctx.SaveUploadedFile(uploadedFile, uploadedFile.Filename); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusCreated, gin.H{"status": "failed", "error": "File is required"})
		return
	}

	quantity, err := strconv.Atoi(ctx.DefaultPostForm("quantity", "0"))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if quantity < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Quantity must be greater than 1 or equal"})
		return
	}

	price, err := strconv.Atoi(ctx.DefaultPostForm("price", "0"))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if price <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Price must be greater than 0"})
		return
	}

	imageFile, err := os.Open(uploadedFile.Filename)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	imageFileName := path.Base(uploadedFile.Filename)

	imageCid, err := ac.web3StorageClient.Put(context.Background(), imageFile)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	imageUrl := fmt.Sprintf("https://%s.ipfs.dweb.link/%s\n", imageCid.String(), imageFileName)

	err = os.Remove(imageFileName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	var token models.Token

	token.Title = ctx.PostForm("title")
	token.Description = ctx.PostForm("description")
	token.CategoryID = ctx.PostForm("category_id")
	token.CollectionID = ctx.PostForm("collection_id")
	token.FractionID = ctx.PostForm("fraction")
	token.Quantity = quantity
	token.LastPrice = price
	token.Image = imageUrl
	token.UpdatedAt = time.Now()
	token.CreatedAt = time.Now()

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

	uriUrl := fmt.Sprintf("https://%s.ipfs.dweb.link/%s\n", uriCid.String(), uriFileName)

	err = os.Remove(uriFileName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	token.Uri = uriUrl

	tokenId, err := ac.repository.InsertToken(token)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	cacheKey := "token-list-*"

	var keys []string
	var cursor uint64
	keys, _, err = ac.redisClient.Scan(cursor, cacheKey, 0).Result()
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

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": gin.H{"tokenId": tokenId}})
}

func (ac *TokenController) GetTokenList(ctx *gin.Context) {
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
	orderBy := ctx.DefaultQuery("order_by", "created_at")
	orderOption := ctx.DefaultQuery("order_option", "ASC")

	var tokens []models.Token

	cacheKey := fmt.Sprintf("token-list-%d-%d-%s-%s-%s", offset, limit, keyword, orderBy, orderOption)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" {
		err := json.Unmarshal([]byte(cache), &tokens)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"tokens": tokens}})
	}

	tokens, err = ac.repository.GetTokenList(limit, offset, keyword, orderBy, orderOption)

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

	if cache != "" {
		err := json.Unmarshal([]byte(cache), &token)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"token": token}})
	}

	token, err = ac.repository.GetTokenData(id)

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

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"token": token}})
}

func (ac *TokenController) UpdateToken(ctx *gin.Context) {
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

	token, err := ac.repository.GetTokenData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if token.Uri == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "Token not found"})
		return
	}

	quantity, err := strconv.Atoi(ctx.DefaultPostForm("quantity", "0"))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if quantity < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Quantity must be greater than 1"})
		return
	}

	token.Title = ctx.DefaultPostForm("title", token.Title)
	token.Description = ctx.DefaultPostForm("description", token.Description)
	token.CategoryID = ctx.DefaultPostForm("category_id", token.CategoryID)
	token.CollectionID = ctx.DefaultPostForm("collection_id", token.CollectionID)
	token.FractionID = ctx.DefaultPostForm("fraction_id", token.FractionID)
	token.UpdatedAt = time.Now()

	if quantity != 0 {
		token.Quantity = quantity
	}

	err = ac.repository.UpdateToken(token.ID, token)

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
	var cursor uint64
	keys, _, err = ac.redisClient.Scan(cursor, cacheKey, 0).Result()
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

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Token has been updated"})
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

	err = ac.repository.DeleteToken(id)

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
	var cursor uint64
	keys, _, err = ac.redisClient.Scan(cursor, cacheKey, 0).Result()
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

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Token has been deleted"})
}
