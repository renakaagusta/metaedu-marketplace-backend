package middlewares

import (
	"metaedu-marketplace/repositories"
	"metaedu-marketplace/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthorizationMiddleware struct {
	userRepository *repositories.UserRepository
	jwtProvider    *utils.JwtHmacProvider
}

func NewAuthorizationMiddleware(userRepository repositories.UserRepository, jwtProvider *utils.JwtHmacProvider) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{&userRepository, jwtProvider}
}

func (ac *AuthorizationMiddleware) VerifyToken(ctx *gin.Context) {
	authorization := ctx.Request.Header.Get("Authorization")

	if authorization == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "error": "User not authorized"})
		ctx.Abort()
		return
	}

	accessToken := strings.Replace(authorization, "Bearer ", "", 1)

	result, err := ac.jwtProvider.Verify(accessToken)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "error": "User not authorized"})
		ctx.Abort()
		return
	}

	userId, err := uuid.Parse(result.Subject)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "error": "User not authorized"})
		ctx.Abort()
		return
	}

	user, err := ac.userRepository.GetUserByID(userId)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "error": "User not authorized"})
		ctx.Abort()
		return
	}

	ctx.Set("user", user)
}
