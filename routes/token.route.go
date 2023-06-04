package routes

import (
	"metaedu-marketplace/controllers"
	"metaedu-marketplace/middlewares"

	"github.com/gin-gonic/gin"
)

type TokenRoutes struct {
	authorizationMiddleware middlewares.AuthorizationMiddleware
	tokenController         controllers.TokenController
}

func NewTokenRoutes(authorizationMiddleware middlewares.AuthorizationMiddleware, tokenController controllers.TokenController) TokenRoutes {
	return TokenRoutes{authorizationMiddleware, tokenController}
}

func (rc *TokenRoutes) TokenRoute(rg *gin.RouterGroup) {

	router := rg.Group("/token")

	router.GET("/:id/transaction", rc.tokenController.GetTokenTransactionList)
	router.GET("/:id", rc.tokenController.GetTokenData)
	router.PUT("/:id", rc.authorizationMiddleware.VerifyToken, rc.tokenController.UpdateToken)
	router.DELETE("/:id", rc.authorizationMiddleware.VerifyToken, rc.tokenController.DeleteToken)
	router.POST("/", rc.authorizationMiddleware.VerifyToken, rc.tokenController.InsertToken)
	router.GET("/", rc.tokenController.GetTokenList)
}
