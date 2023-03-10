package middlewares

import (
	"metaedu-marketplace/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthorizationMiddleware struct {
	jwtProvider *utils.JwtHmacProvider
}

func NewAuthorizationMiddleware(jwtProvider *utils.JwtHmacProvider) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{jwtProvider}
}

func (ac *AuthorizationMiddleware) VerifyToken(ctx *gin.Context) {
	token := ctx.Request.Header.Get("access-token")

	if token == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "error": "User not authorized"})
		ctx.Abort()
		return
	} else {
		ctx.Next()
	}
}
