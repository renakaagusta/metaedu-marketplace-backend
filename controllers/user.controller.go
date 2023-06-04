package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	models "metaedu-marketplace/models"
	"metaedu-marketplace/repositories"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/web3-storage/go-w3s-client"
)

type UserController struct {
	userRepository    *repositories.UserRepository
	web3StorageClient w3s.Client
	redisClient       *redis.Client
}

func NewUserController(userRepository *repositories.UserRepository, web3StorageClient w3s.Client, redisClient *redis.Client) *UserController {
	return &UserController{userRepository, web3StorageClient, redisClient}
}

func (ac *UserController) GetMyUserData(ctx *gin.Context) {
	user, isExist := ctx.Get("user")

	if !isExist {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": "User data is not valid"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"user": user}})
}

func (ac *UserController) GetUserData(ctx *gin.Context) {
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

	var user models.User

	cacheKey := fmt.Sprintf("user-data-%s", id)
	cache, err := ac.redisClient.Get(cacheKey).Result()

	if err != nil && err.Error() != "redis: nil" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if cache != "" && cache != "null" {
		err := json.Unmarshal([]byte(cache), &user)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"user": user}})
		return
	}

	user, err = ac.userRepository.GetUserByID(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if user.Address == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "User not found"})
		return
	}

	cacheBytes, err := json.Marshal(user)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	err = ac.redisClient.Set(cacheKey, cacheBytes, 0).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"user": user}})
}

func (ac *UserController) UpdateUser(ctx *gin.Context) {
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

	uploadedPhoto, err := ctx.FormFile("photo")

	if err == nil {
		if err := ctx.SaveUploadedFile(uploadedPhoto, uploadedPhoto.Filename); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
			return
		}
	}

	uploadedCover, err := ctx.FormFile("cover")

	if err == nil {
		if err := ctx.SaveUploadedFile(uploadedCover, uploadedCover.Filename); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": err.Error()})
			return
		}
	}

	user, err := ac.userRepository.GetUserByID(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	if user.Address == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "User not found"})
		return
	}

	var photoUrl string

	if uploadedPhoto != nil {
		photoFile, err := os.Open(uploadedPhoto.Filename)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		photoFileName := path.Base(uploadedPhoto.Filename)

		photoCid, err := ac.web3StorageClient.Put(context.Background(), photoFile)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}

		photoUrl = fmt.Sprintf("https://%s.ipfs.dweb.link/%s\n", photoCid.String(), photoFileName)

		err = os.Remove(photoFileName)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}
	}

	var coverUrl string

	if uploadedCover != nil {
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

		coverUrl = fmt.Sprintf("https://%s.ipfs.dweb.link/%s\n", coverCid.String(), coverFileName)

		err = os.Remove(coverFileName)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}
	}

	user.Name = ctx.DefaultPostForm("name", user.Name)
	user.Email = ctx.DefaultPostForm("email", user.Email)

	if photoUrl != "" {
		user.Photo = photoUrl
	}

	if coverUrl != "" {
		user.Cover = coverUrl
	}

	user.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	err = ac.userRepository.UpdateUser(user.ID, user)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("user-data-%s", id)
	err = ac.redisClient.Del(cacheKey, idParam).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "User has been updated"})
}
