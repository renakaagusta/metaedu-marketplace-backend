package controllers

import (
	"database/sql"
	"net/http"
	"time"

	models "metaedu-marketplace/models"
	"metaedu-marketplace/repositories"
	"metaedu-marketplace/utils"

	"github.com/gin-gonic/gin"
)

type SignInRequestBody struct {
	Address string `json:"address"`
	Nonce   string `json:"nonce"`
	Sig     string `json:"sig"`
}

type AuthenticationController struct {
	repository  *repositories.UserRepository
	jwtProvider *utils.JwtHmacProvider
}

func NewAuthenticationController(repository *repositories.UserRepository, jwtProvider *utils.JwtHmacProvider) *AuthenticationController {
	return &AuthenticationController{repository, jwtProvider}
}

func (ac *AuthenticationController) SignUp(ctx *gin.Context) {
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
		return
	}

	nonce, err := utils.GetNonce()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	user.Nonce = nonce
	user.Status = "active"
	user.Role = "user"
	user.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	user.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	existingUser := ac.repository.GetUserByAddress(user.Address)

	if (existingUser != models.User{}) {
		ctx.JSON(http.StatusCreated, gin.H{"status": "failed", "error": "User has been registered"})
		return
	}

	userID := ac.repository.InsertUser(user)

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": gin.H{"userId": userID}})
}

func (ac *AuthenticationController) GetUserNonce(ctx *gin.Context) {
	address := ctx.Param("address")

	existingUser := ac.repository.GetUserByAddress(address)

	if (existingUser == models.User{}) {
		ctx.JSON(http.StatusOK, gin.H{"status": "failed", "error": "User doesn't exist"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": gin.H{"nonce": existingUser.Nonce}})
}

func (ac *AuthenticationController) SignIn(ctx *gin.Context) {
	var requestBody SignInRequestBody

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user := ac.repository.GetUserByAddress(requestBody.Address)

	if (user == models.User{}) {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "error": "Address not found"})
		return
	}

	if user.Nonce != requestBody.Nonce {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Nonce is not valid"})
		return
	}

	err := utils.Verify(user.Address, user.Nonce, requestBody.Sig)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Signature is not valid"})
		return
	}

	updatedNonce, err := utils.GetNonce()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": "Failed to create nonce"})
		return
	}

	user.Nonce = updatedNonce

	err = ac.repository.UpdateUser(user.ID, user)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": "Failed to update user nonce"})
		return
	}

	signedToken, err := ac.jwtProvider.CreateJWT(user.ID.String())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": "Failed to create token access"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": gin.H{"access_token": signedToken}})
}
