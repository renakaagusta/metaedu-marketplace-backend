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

type TokenCategoryController struct {
	repository        *repositories.TokenCategoryRepository
	web3StorageClient w3s.Client
	redisClient       *redis.Client
}

func NewTokenCategoryController(repository *repositories.TokenCategoryRepository, web3StorageClient w3s.Client, redisClient *redis.Client) *TokenCategoryController {
	return &TokenCategoryController{repository, web3StorageClient, redisClient}
}

func (ac *TokenCategoryController) InsertTokenCategory(ctx *gin.Context) {
	uploadedIcon, err := ctx.FormFile("icon")

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Image is required"})
		return
	}

	if err := ctx.SaveUploadedFile(uploadedIcon, uploadedIcon.Filename); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	iconFile, err := os.Open(uploadedIcon.Filename)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	iconFileName := path.Base(uploadedIcon.Filename)

	iconCid, err := ac.web3StorageClient.Put(context.Background(), iconFile)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	iconUrl := fmt.Sprintf("https://%s.ipfs.dweb.link/%s\n", iconCid.String(), iconFileName)

	err = os.Remove(iconFileName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	var tokenCategory models.TokenCategory

	tokenCategory.Title = ctx.PostForm("title")
	tokenCategory.Description = ctx.PostForm("description")
	tokenCategory.Icon = iconUrl
	tokenCategory.Status = "active"
	tokenCategory.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	tokenCategory.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	tokenCategoryId, err := ac.repository.InsertTokenCategory(tokenCategory)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	cacheKey := "token-category-list-*"

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

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": gin.H{"token_category_id": tokenCategoryId}})
}

func (ac *TokenCategoryController) GetTokenCategoryList(ctx *gin.Context) {
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

	status := "active"

	var tokenCategories []models.TokenCategory

	cacheKey := fmt.Sprintf("token-category-list-%d-%d-%s-%s-%s", offset, limit, keyword, orderBy, orderOption)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &tokenCategories)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"tokenCategories": tokenCategories}})
		return
	}

	tokenCategories, err = ac.repository.GetTokenCategoryList(offset, limit, keyword, &status, orderBy, orderOption)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheBytes, err := json.Marshal(tokenCategories)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"tokenCategories": tokenCategories}})
}

func (ac *TokenCategoryController) GetTokenCategoryData(ctx *gin.Context) {
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

	var tokenCategory models.TokenCategory

	cacheKey := fmt.Sprintf("token-category-data-%s", id)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &tokenCategory)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"token_category": tokenCategory}})
		return
	}

	tokenCategory, err = ac.repository.GetTokenCategoryData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if tokenCategory.Title == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "Token Category not found"})
		return
	}

	cacheBytes, err := json.Marshal(tokenCategory)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"token_category": tokenCategory}})
}

func (ac *TokenCategoryController) UpdateTokenCategory(ctx *gin.Context) {
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

	tokenCategory, err := ac.repository.GetTokenCategoryData(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	tokenCategory.Title = ctx.DefaultPostForm("title", tokenCategory.Title)
	tokenCategory.Description = ctx.DefaultPostForm("description", tokenCategory.Description)
	tokenCategory.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	err = ac.repository.UpdateTokenCategory(tokenCategory.ID, tokenCategory)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("token-category-data-%s", id)
	err = ac.redisClient.Del(cacheKey, idParam).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey = "token-category-list-*"

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

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Token Category has been updated"})
}

func (ac *TokenCategoryController) DeleteTokenCategory(ctx *gin.Context) {
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

	err = ac.repository.DeleteTokenCategory(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("token-category-data-%s", id)
	err = ac.redisClient.Del(cacheKey, idParam).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey = "token-category-list-*"

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

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Token Category has been deleted"})
}
